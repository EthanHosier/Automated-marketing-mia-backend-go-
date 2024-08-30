package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethanhosier/mia-backend-go/prompts"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/sashabaranov/go-openai"
)

const (
	tokensFilePath = "canva-tokens.json"
)

var (
	mu sync.Mutex
)

func loadTokens() (*types.Token, error) {
	file, err := os.ReadFile(tokensFilePath)
	if err != nil {
		return nil, err
	}

	var tokens types.Token
	err = json.Unmarshal(file, &tokens)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func AccessToken() (string, error) {
	mu.Lock()

	tokens, err := loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %v", err)
	}

	if tokens.ExpiresIn > time.Now().Unix() {
		// Token is still valid
		mu.Unlock()
		return tokens.AccessToken, nil
	}

	mu.Unlock()
	return refreshAccessToken()
}

func saveTokens(tokens *types.Token) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(tokensFilePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func refreshAccessToken() (string, error) {
	mu.Lock()
	defer mu.Unlock()

	tokens, err := loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %v", err)
	}

	if tokens.ExpiresIn > time.Now().Unix() {
		// Token is still valid
		return tokens.AccessToken, nil
	}

	if tokens.RefreshToken == "" {
		return "", fmt.Errorf("refresh token not found")
	}

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", tokens.RefreshToken)

	req, err := http.NewRequest("POST", "https://api.canva.com/rest/v1/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	clientID := os.Getenv("CANVA_CLIENT_ID")
	clientSecret := os.Getenv("CANVA_CLIENT_SECRET")
	if clientID == "" {
		return "", fmt.Errorf("CANVA_CLIENT_ID not set")
	}
	if clientSecret == "" {
		return "", fmt.Errorf("CANVA_CLIENT_SECRET not set")
	}

	req.URL.RawQuery = form.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error refreshing access token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error refreshing access token: %s", body)
	}

	var newTokens types.Token
	err = json.NewDecoder(resp.Body).Decode(&newTokens)

	if err != nil {
		return "", err
	}

	// Update the expiration time and save the tokens
	newTokens.ExpiresIn = time.Now().Unix() + (newTokens.ExpiresIn - time.Now().Unix())
	err = saveTokens(&newTokens)
	if err != nil {
		return "", err
	}

	log.Println("Canva token refreshed successfully")
	return newTokens.AccessToken, nil
}

func StartCanvaTokenRefresher(interval time.Duration) {
	_, err := refreshAccessToken()
	if err != nil {
		log.Fatalf("Error refreshing token: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		_, err := refreshAccessToken()
		if err != nil {
			log.Printf("Error refreshing token: %v", err)
		} else {
			log.Println("Token refreshed successfully")
		}
	}
}

func PickBestImages(candidateImages []string, campaignInfo string, imageFields []types.TemplateFields, llmClient *LLMClient) ([]string, error) {
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
				bestImagePairChan <- BestImagePair{i, 0}
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
		return fmt.Errorf("error decoding response body:", err)
	}

	var jobStatusResponse types.JobStatus
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
