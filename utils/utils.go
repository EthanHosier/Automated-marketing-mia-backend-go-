package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	open, close := getBrackets(b)

	start := strings.Index(jsonString, open)
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(jsonString, close)
	if end == -1 || end <= start {
		return ""
	}

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

func PopulateTemplate(nearestTemplate types.NearestTemplateResponse, populatedTemplate types.PopulatedTemplate) error {
	populatedTemplateFieldMap := map[string]string{}

	for _, field := range populatedTemplate.Fields {
		populatedTemplateFieldMap[field.Name] = field.Value
	}

	inputData := map[string]map[string]string{}

	for _, field := range nearestTemplate.Fields {
		if field.Type != "text" {
			continue
		}

		inputData[field.Name] = map[string]string{
			"type": "text",
			"text": populatedTemplateFieldMap[field.Name],
		}
	}

	requestData := map[string]interface{}{
		"brand_template_id": nearestTemplate.ID,
		"data":              inputData,
	}

	accessToken, err := AccessToken()
	if err != nil {
		return err
	}

	url := "https://api.canva.com/rest/v1/autofills"
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("Error marshalling request data:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)

	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received non-OK response: %s\n", resp.Status)
	}

	var responseBody types.UpdateTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("Error decoding response body:", err)
	}

	type JobStatus struct {
		Job struct {
			ID     string `json:"id"`
			Result struct {
				Type   string `json:"type"`
				Design struct {
					CreatedAt int64  `json:"created_at"` // Use int64 for Unix timestamp
					ID        string `json:"id"`
					Title     string `json:"title"`
					UpdatedAt int64  `json:"updated_at"` // Use int64 for Unix timestamp
					Thumbnail struct {
						URL string `json:"url"`
					} `json:"thumbnail"`
					URL  string `json:"url"`
					URLs struct {
						EditURL string `json:"edit_url"`
						ViewURL string `json:"view_url"`
					} `json:"urls"`
				} `json:"design"`
			} `json:"result"`
			Status string `json:"status"`
		} `json:"job"`
	}

	var jobStatusResponse JobStatus
	for jobStatusResponse.Job.Status != "success" {
		fmt.Println("Checking job status...")
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("https://api.canva.com/rest/v1/autofills/%s", responseBody.Job.ID)
		req, err := http.NewRequest("GET", statusURL, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)

		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %vs", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-OK response: %s\n", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&jobStatusResponse); err != nil {
			return fmt.Errorf("error decoding response body: %v", err)

		}
		fmt.Printf("Job status: %+v\n", jobStatusResponse)
	}
	fmt.Printf("Job status: %+v\n", jobStatusResponse)

	fmt.Printf("Response: %+v\n", responseBody)
	return nil
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
