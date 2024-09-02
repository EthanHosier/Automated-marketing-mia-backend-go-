package utils

import (
	"bytes"
	"encoding/base64"
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
	// I CHANGED THS FROM "	newTokens.ExpiresIn = time.Now().Unix() + (newTokens.ExpiresIn - time.Now().Unix())""
	newTokens.ExpiresIn = time.Now().Unix() + newTokens.ExpiresIn // now absolute time of expiration
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
		}
	}
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

func PopulateTemplate(ID string, imageFields []types.PopulatedField, textFields []types.PopulatedField, colorFields []types.PopulatedColorField) (*types.UpdateTemplateResult, error) {
	inputData := map[string]map[string]string{}

	for _, field := range imageFields {
		inputData[field.Name] = map[string]string{
			"type":     "image",
			"asset_id": field.Value,
		}
	}

	for _, field := range textFields {
		inputData[field.Name] = map[string]string{
			"type": "text",
			"text": field.Value,
		}
	}

	for _, field := range colorFields {
		inputData[field.Name] = map[string]string{
			"type":     "image",
			"asset_id": field.Color,
		}
	}

	log.Printf("Input data: %+v\n", inputData)

	requestData := map[string]interface{}{
		"brand_template_id": ID,
		"data":              inputData,
	}

	accessToken, err := AccessToken()
	if err != nil {
		return nil, err
	}

	url := "https://api.canva.com/rest/v1/autofills"
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)

	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received non-OK response: %s\n", resp.Status)
	}

	var responseBody types.UpdateTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("error decoding response body:", err)
	}

	var jobStatusResponse types.UpdateTemplateJobStatus
	for jobStatusResponse.Job.Status != "success" && jobStatusResponse.Job.Status != "failed" {
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("https://api.canva.com/rest/v1/autofills/%s", responseBody.Job.ID)
		req, err := http.NewRequest("GET", statusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)

		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %vs", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-OK response: %s\n", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&jobStatusResponse); err != nil {
			return nil, fmt.Errorf("error decoding response body: %v", err)
		}

		log.Println("Template update status:", jobStatusResponse.Job.Status)
	}

	if jobStatusResponse.Job.Status == "failed" {
		return nil, fmt.Errorf("job failed")
	}

	return &jobStatusResponse.Job.Result, nil
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

func uploadAsset(asset []byte, name string) (*types.UploadAssetResponse, error) {
	accessToken, err := AccessToken()
	if err != nil {
		log.Fatalf("Error getting access token: %v", err)
	}

	url := "https://api.canva.com/rest/v1/asset-uploads"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(asset))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	metadata := map[string]string{
		"name_base64": base64.StdEncoding.EncodeToString([]byte(name)),
	}
	metadataJSON, err := json.Marshal(metadata)

	if err != nil {
		log.Fatalf("Error marshalling metadata: %v", err)
	}

	req.Header = http.Header{
		"Authorization":         {"Bearer " + accessToken},
		"Content-Type":          {"application/octet-stream"},
		"Asset-Upload-Metadata": {string(metadataJSON)},
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("received non-OK response: %s\n", body)
	}

	var uploadAssetResponse types.UploadAssetResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadAssetResponse); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}

	for uploadAssetResponse.Job.Status != "success" && uploadAssetResponse.Job.Status != "failed" {
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("https://api.canva.com/rest/v1/asset-uploads/%s", uploadAssetResponse.Job.ID)
		req, err := http.NewRequest("GET", statusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&uploadAssetResponse); err != nil {
			return nil, fmt.Errorf("error decoding response body: %v", err)
		}
		log.Println("Asset upload status:", uploadAssetResponse.Job.Status)
	}

	if uploadAssetResponse.Job.Status == "failed" {
		return nil, fmt.Errorf("asset upload failed")
	}

	return &uploadAssetResponse, nil
}

func UploadColorAssets(colorFields []types.PopulatedColorField) ([]types.PopulatedColorField, error) {
	colorFieldCh := make(chan types.PopulatedColorField, len(colorFields))
	colorFieldWg := sync.WaitGroup{}
	colorFieldWg.Add(len(colorFields))

	for _, field := range colorFields {
		go func(field types.PopulatedColorField) {
			defer colorFieldWg.Done()

			colorImg, err := CreateColorImage(field.Color)
			if err != nil {
				log.Println("error creating color image: ", err)
				return
			}

			resp, err := uploadAsset(colorImg, "name")
			if err != nil {
				log.Println("error uploading color image: ", err)
				return
			}

			field.Color = resp.Job.ID
			colorFieldCh <- field
		}(field)
	}

	colorFieldWg.Wait()
	close(colorFieldCh)

	colorFields = []types.PopulatedColorField{}
	for field := range colorFieldCh {
		colorFields = append(colorFields, field)
	}

	return colorFields, nil
}

func UploadImageAssets(imageFields []types.PopulatedField, bestImages []string) ([]types.PopulatedField, error) {
	imageFieldsCh := make(chan types.PopulatedField, len(imageFields))
	imageFieldsWg := sync.WaitGroup{}
	imageFieldsWg.Add(len(imageFields))

	for i, field := range imageFields {
		go func(field types.PopulatedField, i int) {
			defer imageFieldsWg.Done()
			img, err := DownloadImage(bestImages[i])
			if err != nil {
				log.Println("error downloading image: ", err)
				return
			}

			resp, err := uploadAsset(img, "name")
			if err != nil {
				log.Println("error uploading image: ", err)
				return
			}

			field.Value = resp.Job.ID
			imageFieldsCh <- field
		}(field, i)
	}

	imageFieldsWg.Wait()
	close(imageFieldsCh)

	imageFields = []types.PopulatedField{}
	for field := range imageFieldsCh {
		imageFields = append(imageFields, field)
	}

	return imageFields, nil
}
