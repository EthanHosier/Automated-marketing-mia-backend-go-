package canva

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	net_http "net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/utils"
)

const (
	tokenEndpoint        = "https://api.canva.com/rest/v1/oauth/token"
	autofillEndpoint     = "https://api.canva.com/rest/v1/autofills"
	assetUploadsEndpoint = "https://api.canva.com/rest/v1/asset-uploads"
)

type CanvaClient interface {
	PopulateTemplate(ID string, imageFields []ImageField, textFields []TextField, colorFields []ColorField) (*UpdateTemplateResult, error)
	UploadImageAssets(images []string) ([]string, error)
	UploadColorAssets(colors []string) ([]string, error)
}

type CanvaHttpClient struct {
	clientID     string
	clientSecret string
	httpClient   http.Client

	tokensFilePath  string
	mu              sync.Mutex
	tokenBufferSecs int
}

func NewClient(clientID string, clientSecret string, tokensFilePath string, httpClient http.Client, tokenBufferSecs int) *CanvaHttpClient {
	canvaClient := CanvaHttpClient{
		clientID:        clientID,
		clientSecret:    clientSecret,
		httpClient:      httpClient,
		tokensFilePath:  tokensFilePath,
		mu:              sync.Mutex{},
		tokenBufferSecs: tokenBufferSecs,
	}

	go canvaClient.startCanvaTokenRefresher(30 * time.Minute)
	return &canvaClient
}

func (c *CanvaHttpClient) loadTokens() (*Tokens, error) {
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
func (c *CanvaHttpClient) accessToken() (string, error) {
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

func (c *CanvaHttpClient) saveTokens(tokens *Tokens) error {
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

func (c *CanvaHttpClient) refreshAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tokens, err := c.loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %v", err)
	}

	if tokens.ExpiresIn > (time.Now().Unix() + int64(c.tokenBufferSecs)) {
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

	if newTokens.AccessToken == "" {
		return "", fmt.Errorf("access token not found in response")
	}

	if newTokens.RefreshToken == "" {
		return "", fmt.Errorf("refresh token not found in response")
	}

	err = c.saveTokens(newTokens)
	if err != nil {
		return "", err
	}

	slog.Info("Canva token refreshed successfully")
	return newTokens.AccessToken, nil
}

func (c *CanvaHttpClient) sendRefreshAccessTokenRequest(refreshToken string) (*Tokens, error) {
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", refreshToken)

	req, err := c.httpClient.NewRequest("POST", tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = form.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
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

func (c *CanvaHttpClient) startCanvaTokenRefresher(interval time.Duration) {
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

func (c *CanvaHttpClient) PopulateTemplate(ID string, imageFields []ImageField, textFields []TextField, colorFields []ColorField) (*UpdateTemplateResult, error) {
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

func (c *CanvaHttpClient) sendAutofillRequest(requestData map[string]interface{}) (*net_http.Response, error) {
	accessToken, err := c.accessToken()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data: %v", err)
	}

	req, err := c.httpClient.NewRequest("POST", autofillEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)

	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *CanvaHttpClient) decodeUpdateTemplateResult(resp *net_http.Response) (*UpdateTemplateResult, error) {
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

func (c *CanvaHttpClient) decodeUpdateTemplateJobResult(jobID string) (*UpdateTemplateResult, error) {
	var jobStatusResponse UpdateTemplateJobStatus
	for jobStatusResponse.Job.Status != "success" && jobStatusResponse.Job.Status != "failed" {
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("%s/%s", autofillEndpoint, jobID)
		req, err := c.httpClient.NewRequest("GET", statusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		accessToken, err := c.accessToken()
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

		// slog.Info("Decoded update template response", "updateTemplateResponseStatus", jobStatusResponse.Job.Status)
	}

	if jobStatusResponse.Job.Status == "failed" {
		return nil, fmt.Errorf("job failed")
	}

	return &jobStatusResponse.Job.Result, nil
}

func (c *CanvaHttpClient) uploadAsset(asset []byte, name string) (*Asset, error) {
	resp, err := c.sendUploadAssetRequest(asset, name)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	return c.decodeUploadAssetResponse(resp)
}

func (c *CanvaHttpClient) sendUploadAssetRequest(asset []byte, name string) (*net_http.Response, error) {
	req, err := c.httpClient.NewRequest("POST", assetUploadsEndpoint, bytes.NewBuffer(asset))
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

	accessToken, err := c.accessToken()
	if err != nil {
		return nil, fmt.Errorf("error getting access token: %v", err)
	}

	req.Header = net_http.Header{
		"Authorization":         {"Bearer " + accessToken},
		"Content-Type":          {"application/octet-stream"},
		"Asset-Upload-Metadata": {string(metadataJSON)},
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("received non-OK response: %s", body)
	}

	return resp, nil
}

func (c *CanvaHttpClient) decodeUploadAssetResponse(resp *net_http.Response) (*Asset, error) {
	defer resp.Body.Close()

	var uploadAssetResponse UploadAssetResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadAssetResponse); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}

	for uploadAssetResponse.Job.Status != "success" && uploadAssetResponse.Job.Status != "failed" {
		time.Sleep(2 * time.Second) // Wait for 2 seconds before checking status

		statusURL := fmt.Sprintf("%s/%s", assetUploadsEndpoint, uploadAssetResponse.Job.ID)
		req, err := c.httpClient.NewRequest("GET", statusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		accessToken, err := c.accessToken()
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

		// slog.Info("Decoded upload asset response", "AssetUploadStatus", uploadAssetResponse.Job.Status)
	}

	if uploadAssetResponse.Job.Status == "failed" {
		return nil, fmt.Errorf("asset upload failed")
	}

	return &uploadAssetResponse.Job.Asset, nil
}

func (c *CanvaHttpClient) UploadColorAssets(colors []string) ([]string, error) {
	tasks := utils.DoAsyncList(colors, func(color string) (string, error) {
		asset, err := c.createAndUploadColorAsset(color)
		if err != nil {
			return "", fmt.Errorf("error creating and uploading color asset: %v", err)
		}

		return asset.ID, nil
	})

	return utils.GetAsyncList(tasks)
}

func (c *CanvaHttpClient) createAndUploadColorAsset(color string) (*Asset, error) {
	colorImg, err := createColorImage(color)
	if err != nil {
		return nil, fmt.Errorf("error creating color image: %v", err)
	}

	return c.uploadAsset(colorImg, "name")
}

func (c *CanvaHttpClient) UploadImageAssets(images []string) ([]string, error) {
	tasks := utils.DoAsyncList(images, func(image string) (string, error) {
		asset, err := c.downloadAndUploadImageAsset(image)
		if err != nil {
			return "", fmt.Errorf("error downloading and uploading image asset: %v", err)
		}

		return asset.ID, nil
	})

	return utils.GetAsyncList(tasks)
}

func (c *CanvaHttpClient) downloadAndUploadImageAsset(image string) (*Asset, error) {
	if strings.HasPrefix(image, "data:") {
		b64data := image[strings.IndexByte(image, ',')+1:]
		imageBytes, err := base64.StdEncoding.DecodeString(b64data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 image: %v", err)
		}

		return c.uploadAsset(imageBytes, "name")
	}

	img, err := c.downloadImage(image)

	if err != nil {
		return nil, fmt.Errorf("error downloading image: %v", err)
	}

	return c.uploadAsset(img, "name")
}

func (c *CanvaHttpClient) downloadImage(imageURL string) ([]byte, error) {
	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("error getting image %s:  %v", imageURL, err)
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
