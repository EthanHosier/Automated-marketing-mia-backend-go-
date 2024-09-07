package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/ethanhosier/mia-backend-go/types"
)

type BracketType int

const (
	SquareBracket BracketType = iota
	CurlyBracket
)

func Round2Dec(val float64) float64 {
	return math.Round(val*100) / 100
}

func RemoveDuplicates(input []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, str := range input {
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}
			result = append(result, str)
		}
	}

	return result
}

func VectorToPostgresFormat(v types.Vector) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func VectorFromPostgresFormat(s string) (types.Vector, error) {
	var v types.Vector
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func minMax(nums []float32) (float32, float32) {
	min, max := nums[0], nums[0]
	for _, num := range nums {
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}
	return min, max
}

func Normalize(x float32, nums []float32) float32 {
	min, max := minMax(nums)
	return float32(x-min) / float32(max-min)
}

func PageTextContents(url string) (string, error) {
	endpoint := SinglePageBodyTextScraperUrl + url

	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var response types.SinglePageBodyTextScraperResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

func CleanText(text string) string {
	// Trim leading and trailing whitespace
	text = strings.TrimSpace(text)

	// Replace multiple spaces, newlines, and tabs with a single space
	text = strings.Join(strings.Fields(text), " ")

	return text
}

func FirstNumberInString(s string) (int, error) {
	num := ""
	for _, char := range s {
		if unicode.IsDigit(char) {
			num += string(char)
		} else if len(num) > 0 {
			// If we've found a number and hit a non-digit character, break out
			break
		}
	}

	// If no digits were found, return an error
	if num == "" {
		return 0, errors.New("no number found in the string")
	}

	// Convert the string of digits to an integer
	number, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func DownloadImage(imageURL string) ([]byte, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: %s", resp.Status)
	}

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return imgData, nil
}

// ValidateMapKeys checks if the keys in the map match the JSON tags of the struct fields.
func ValidateMapKeys[T any](inputStruct T, inputMap map[string]interface{}) error {
	// Get the type of the struct
	structType := reflect.TypeOf(inputStruct)

	// Create a set of valid keys from the struct's JSON tag names
	validKeys := make(map[string]struct{})
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")

		// Handle cases where the json tag is omitted or has specific options like `json:"-"`
		if jsonTag == "" {
			jsonTag = field.Name
		} else if jsonTag == "-" {
			continue // Skip fields with `json:"-"`
		} else {
			// In case the tag contains options, e.g., `json:"name,omitempty"`
			jsonTag = parseJSONTag(jsonTag)
		}
		validKeys[jsonTag] = struct{}{}
	}

	// Check if all keys in the map are valid JSON field names
	for key := range inputMap {
		if _, exists := validKeys[key]; !exists {
			return fmt.Errorf("invalid key: %s", key)
		}
	}

	return nil
}

func parseJSONTag(tag string) string {
	if commaIndex := len(tag); commaIndex != -1 {
		return tag[:commaIndex]
	}
	return tag
}

// TODO: MAKE THIS TAKE INTO CONSIDERATION THE PROMPT OF THE IMAGE FIELD BEFORE POPULATING (types.TemplateField.description)
func PickBestImages(candidateImages []string, campaignInfo string, imageFields []types.PopulatedField, llmClient *LLMClient) ([]string, error) {
	bestImagesWg := sync.WaitGroup{}

	if len(candidateImages) > 50 {
		return nil, fmt.Errorf("> 50 candidate images")
	}

	if len(candidateImages) == 0 {
		return nil, fmt.Errorf("no candidate images supplied")
	}

	type BestImagePair struct {
		fieldIndex     int
		bestImageIndex int
	}

	bestImagePairChan := make(chan BestImagePair, len(imageFields))
	bestImagesWg.Add(len(imageFields))

	for i, field := range imageFields {
		prompt := prompts.PickBestImagePrompt(campaignInfo, field)
		go func(prompt string, i int) {
			defer bestImagesWg.Done()

			bestImage, err := llmClient.OpenaiImageCompletion(prompt, candidateImages, openai.GPT4o)
			if err != nil {
				log.Printf("Error getitng openai image completion: %v", err)
				log.Printf("candidate images: %v", candidateImages)
				bestImagePairChan <- BestImagePair{i, 0}
				return
			}

			index, err := FirstNumberInString(bestImage)
			if err != nil {
				log.Printf("Error converting number to string: %v", err)
				bestImagePairChan <- BestImagePair{i, 0}
			}
			bestImagePairChan <- BestImagePair{i, index}
		}(prompt, i)
	}

	bestImagesWg.Wait()
	close(bestImagePairChan)

	bestImages := make([]string, len(imageFields))
	for pair := range bestImagePairChan {
		bestImages[pair.fieldIndex] = candidateImages[pair.bestImageIndex]
	}

	return bestImages, nil
}

func IsValidImageURL(url string) bool {
	if !strings.HasPrefix(url, "http") {
		return false
	}

	// Common image file extensions
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(url), ext) {
			return true
		}
	}
	return false
}
