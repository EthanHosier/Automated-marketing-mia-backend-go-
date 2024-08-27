package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

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
		return fmt.Errorf("Error creating request:", err)

	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received non-OK response: %s\n", resp.Status)
	}

	var responseBody types.UpdateTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("Error decoding response body:", err)
	}

	var jobStatusResponse map[string]interface{}
	for {
		fmt.Println("Checking job status...")
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("https://api.canva.com/rest/v1/autofills/%s", responseBody.Job.ID)
		req, err := http.NewRequest("GET", statusURL, nil)
		if err != nil {
			return fmt.Errorf("Error creating request:", err)

		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Error sending request:", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Received non-OK response: %s\n", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&jobStatusResponse); err != nil {
			return fmt.Errorf("Error decoding response body:", err)

		}
		fmt.Printf("Job status: %+v", jobStatusResponse)

		if jobStatusResponse["status"] == "success" {
			break
		}
	}
	fmt.Printf("Job status: %+v", jobStatusResponse)

	fmt.Printf("Response: %+v\n", responseBody)
	return nil
}

func IsValidImageURL(url string) bool {
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
