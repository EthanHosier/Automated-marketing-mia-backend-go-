package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/ethanhosier/mia-backend-go/types"
)

type BracketType int

const (
	SquareBracket BracketType = iota // Starts from 0
	CurlyBracket                     // Increments to 1
)

func Round2Dec(val float64) float64 {
	return math.Round(val*100) / 100
}

func ExtractJsonObj(jsonString string, b BracketType) string {
	// Remove all newline characters from the input string
	jsonString = strings.ReplaceAll(jsonString, "\n", "")

	// Get the opening and closing brackets for the specific type
	open, close := getBrackets(b)

	// Find the start index of the opening bracket
	start := strings.Index(jsonString, open)
	if start == -1 {
		return ""
	}

	// Find the last index of the closing bracket
	end := strings.LastIndex(jsonString, close)
	if end == -1 || end <= start {
		return ""
	}

	// Return the substring containing the JSON object
	return jsonString[start : end+1]
}

func getBrackets(b BracketType) (string, string) {
	if b == SquareBracket {
		return "[", "]"
	}

	return "{", "}"
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

func StringifyWebsiteData(data types.WebsiteData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title: %s\n", data.Title))
	sb.WriteString(fmt.Sprintf("Meta Description: %s\n", data.MetaDescription))

	sb.WriteString("Headings:\n")
	for key, values := range data.Headings {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", key, strings.Join(values, ", ")))
	}

	sb.WriteString(fmt.Sprintf("Keywords: %s\n", data.Keywords))
	sb.WriteString(fmt.Sprintf("Links: %s\n", strings.Join(data.Links, ", ")))
	sb.WriteString(fmt.Sprintf("Summary: %s\n", data.Summary))
	sb.WriteString(fmt.Sprintf("Categories: %s\n", strings.Join(data.Categories, ", ")))

	return sb.String()
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

func HexToColor(hex string) (color.Color, error) {
	if strings.HasPrefix(hex, "#") {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color format")
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, err
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
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
