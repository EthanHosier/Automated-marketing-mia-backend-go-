package canva

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	tokenEndpoint        = "https://api.canva.com/rest/v1/oauth/token"
	autofillEndpoint     = "https://api.canva.com/rest/v1/autofills"
	assetUploadsEndpoint = "https://api.canva.com/rest/v1/asset-uploads"
)

type CanvaClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client // remember to set timeout as requestTimeout in new

	tokensFilePath string
	mu             sync.Mutex
}

func New(clientID string, clientSecret string, tokensFilePath string, httpClient *http.Client) *CanvaClient {
	canvaClient := CanvaClient{
		clientID:       clientID,
		clientSecret:   clientSecret,
		httpClient:     httpClient,
		tokensFilePath: tokensFilePath,
		mu:             sync.Mutex{},
	}

	go canvaClient.startCanvaTokenRefresher(30 * time.Minute)
	return &canvaClient
}

func (c *CanvaClient) loadTokens() (*Tokens, error) {
	file, err := os.ReadFile(c.tokensFilePath)
	if err != nil {
		return nil, err
	}

	var tokens Tokens
	err = json.Unmarshal(file, &tokens)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

// todo: make this a read write lock???
func (c *CanvaClient) AccessToken() (string, error) {
	c.mu.Lock()

	tokens, err := c.loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %v", err)
	}

	if tokens.ExpiresIn > time.Now().Unix() {
		// Token is still valid
		c.mu.Unlock()
		return tokens.AccessToken, nil
	}

	c.mu.Unlock()
	return c.refreshAccessToken()
}

func (c *CanvaClient) saveTokens(tokens *Tokens) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(c.tokensFilePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *CanvaClient) refreshAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokens, err := c.loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %v", err)
	}

	if tokens.ExpiresIn > (time.Now().Unix() + 300) {
		// Token is still valid with a 5-minute buffer
		return tokens.AccessToken, nil
	}

	if tokens.RefreshToken == "" {
		return "", fmt.Errorf("refresh token not found")
	}

	newTokens, err := c.sendRefreshAccessTokenRequest(tokens.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh access token: %v", err)
	}

	err = c.saveTokens(newTokens)
	if err != nil {
		return "", err
	}

	slog.Info("Canva token refreshed successfully")
	return newTokens.AccessToken, nil
}

func (c *CanvaClient) sendRefreshAccessTokenRequest(refreshToken string) (*Tokens, error) {
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = form.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", body)
	}

	var newTokens Tokens
	err = json.NewDecoder(resp.Body).Decode(&newTokens)

	if err != nil {
		return nil, err
	}

	newTokens.ExpiresIn = time.Now().Unix() + newTokens.ExpiresIn
	return &newTokens, nil
}

func (c *CanvaClient) startCanvaTokenRefresher(interval time.Duration) {
	_, err := c.refreshAccessToken()
	if err != nil {
		slog.Error("Error refreshing token", "error", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		_, err := c.refreshAccessToken()
		if err != nil {
			slog.Info("Error refreshing token", "error", err)
		}
	}
}

func (c *CanvaClient) PopulateTemplate(ID string, imageFields []ImageField, textFields []TextField, colorFields []ColorField) (*UpdateTemplateResult, error) {
	inputData := populateTemplateInputData(imageFields, textFields, colorFields)
	slog.Info("InputData populated", "Input data", inputData)

	requestData := map[string]interface{}{
		"brand_template_id": ID,
		"data":              inputData,
	}

	resp, err := c.sendAutofillRequest(requestData)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	return c.decodeUpdateTemplateResult(resp)
}

func (c *CanvaClient) sendAutofillRequest(requestData map[string]interface{}) (*http.Response, error) {
	accessToken, err := c.AccessToken()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data: %v", err)
	}

	req, err := http.NewRequest("POST", autofillEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)

	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *CanvaClient) decodeUpdateTemplateResult(resp *http.Response) (*UpdateTemplateResult, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	var responseBody UpdateTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}

	return c.decodeUpdateTemplateJobResult(responseBody.Job.ID)
}

func (c *CanvaClient) decodeUpdateTemplateJobResult(jobID string) (*UpdateTemplateResult, error) {
	var jobStatusResponse UpdateTemplateJobStatus
	for jobStatusResponse.Job.Status != "success" && jobStatusResponse.Job.Status != "failed" {
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("%s/%s", autofillEndpoint, jobID)
		req, err := http.NewRequest("GET", statusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)

		}

		accessToken, err := c.AccessToken()
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := c.httpClient.Do(req)

		if err != nil {
			return nil, fmt.Errorf("error sending request: %vs", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&jobStatusResponse); err != nil {
			return nil, fmt.Errorf("error decoding response body: %v", err)
		}

		slog.Info("Decoded update template response", "updateTemplateResponseStatus", jobStatusResponse.Job.Status)

	}

	if jobStatusResponse.Job.Status == "failed" {
		return nil, fmt.Errorf("job failed")
	}

	return &jobStatusResponse.Job.Result, nil
}

func (c *CanvaClient) uploadAsset(asset []byte, name string) (*Asset, error) {
	resp, err := c.sendUploadAssetRequest(asset, name)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	return c.decodeUploadAssetResponse(resp)
}

func (c *CanvaClient) sendUploadAssetRequest(asset []byte, name string) (*http.Response, error) {
	req, err := http.NewRequest("POST", assetUploadsEndpoint, bytes.NewBuffer(asset))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	metadata := map[string]string{
		"name_base64": base64.StdEncoding.EncodeToString([]byte(name)),
	}
	metadataJSON, err := json.Marshal(metadata)

	if err != nil {
		return nil, fmt.Errorf("error marshalling metadata: %v", err)
	}

	accessToken, err := c.AccessToken()
	if err != nil {
		return nil, fmt.Errorf("error getting access token: %v", err)
	}

	req.Header = http.Header{
		"Authorization":         {"Bearer " + accessToken},
		"Content-Type":          {"application/octet-stream"},
		"Asset-Upload-Metadata": {string(metadataJSON)},
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("received non-OK response: %s", body)
	}

	return resp, nil
}

func (c *CanvaClient) decodeUploadAssetResponse(resp *http.Response) (*Asset, error) {
	var uploadAssetResponse UploadAssetResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadAssetResponse); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}

	for uploadAssetResponse.Job.Status != "success" && uploadAssetResponse.Job.Status != "failed" {
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("%s/%s", assetUploadsEndpoint, uploadAssetResponse.Job.ID)
		req, err := http.NewRequest("GET", statusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		accessToken, err := c.AccessToken()
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := c.httpClient.Do(req)
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

		slog.Info("Decoded upload asset response", "AssetUploadStatus", uploadAssetResponse.Job.Status)
	}

	if uploadAssetResponse.Job.Status == "failed" {
		return nil, fmt.Errorf("asset upload failed")
	}

	return &uploadAssetResponse.Job.Asset, nil
}

func (c *CanvaClient) UploadColorAssets(colors []string) ([]string, error) {
	errorCh := make(chan error, len(colors))
	colorIds := make([]string, len(colors))

	mu := sync.Mutex{}

	colorWg := sync.WaitGroup{}
	colorWg.Add(len(colors))

	for i, color := range colors {
		go func(i int, color string) {
			defer colorWg.Done()

			asset, err := c.createAndUploadColorAsset(color)
			if err != nil {
				errorCh <- fmt.Errorf("error creating and uploading image asset: %v", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			colorIds[i] = asset.ID
		}(i, color)
	}

	colorWg.Wait()
	close(errorCh)

	select {
	case err := <-errorCh:
		return nil, err
	default:
		return colorIds, nil
	}
}

func (c *CanvaClient) createAndUploadColorAsset(color string) (*Asset, error) {
	colorImg, err := createColorImage(color)
	if err != nil {
		return nil, fmt.Errorf("error creating color image: %v", err)
	}

	return c.uploadAsset(colorImg, "name")
}

func (c *CanvaClient) UploadImageAssets(images []string) ([]string, error) {
	errorCh := make(chan error, len(images))
	imageIds := make([]string, len(images))

	mu := sync.Mutex{}

	imageIdsWg := sync.WaitGroup{}
	imageIdsWg.Add(len(imageIds))

	for i, field := range images {
		go func(i int, image string) {
			defer imageIdsWg.Done()

			asset, err := c.downloadAndUploadImageAsset(image)
			if err != nil {
				errorCh <- fmt.Errorf("error downloading and uploading image asset: %v", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			imageIds[i] = asset.ID
		}(i, field)
	}

	imageIdsWg.Wait()
	close(errorCh)

	select {
	case err := <-errorCh:
		return nil, err
	default:
		return imageIds, nil
	}
}

func (c *CanvaClient) downloadAndUploadImageAsset(image string) (*Asset, error) {
	img, err := c.downloadImage(image)

	if err != nil {
		return nil, fmt.Errorf("error downloading image: %v", err)
	}

	return c.uploadAsset(img, "name")
}

func (c *CanvaClient) downloadImage(imageURL string) ([]byte, error) {
	resp, err := c.httpClient.Get(imageURL)
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
