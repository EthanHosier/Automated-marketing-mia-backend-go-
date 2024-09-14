package canva

import (
	"testing"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/stretchr/testify/assert"
)

const (
	testTokenBufferSecs = 9999999999
)

func TestCanvaClient_PopulateTemplate(t *testing.T) {
	// given
	var (
		mockClient  = &http.MockHttpClient{}
		canvaClient = NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient, testTokenBufferSecs)

		templateResult = &UpdateTemplateResult{
			Type: "template_update",
			Design: Design{
				CreatedAt: 1694640000,
				ID:        "design_67890",
				Title:     "Spring Collection",
				UpdatedAt: 1694643600,
				Thumbnail: struct {
					URL string `json:"url"`
				}{
					URL: "https://example.com/thumbnail.jpg",
				},
				URL: "https://example.com/design/67890",
				URLs: struct {
					EditURL string `json:"edit_url"`
					ViewURL string `json:"view_url"`
				}{
					EditURL: "https://example.com/edit/67890",
					ViewURL: "https://example.com/view/67890",
				},
			},
		}

		imageFields = []ImageField{}
		textFields  = []TextField{}
		colorFields = []ColorField{}
	)

	mockClient.WillReturnBody("POST", autofillEndpoint, `{"job": {"id": "1234"}}`)
	mockClient.WillReturnBody("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("GET", autofillEndpoint+"/1234", `{"job": {
  "id": "job_12345",
  "result": {
    "type": "template_update",
    "design": {
      "created_at": 1694640000,
      "id": "design_67890",
      "title": "Spring Collection",
      "updated_at": 1694643600,
      "thumbnail": {
        "url": "https://example.com/thumbnail.jpg"
      },
      "url": "https://example.com/design/67890",
      "urls": {
        "edit_url": "https://example.com/edit/67890",
        "view_url": "https://example.com/view/67890"
      }
    }
  },
  "status": "success"
}}`)

	// when
	result, err := canvaClient.PopulateTemplate("testTemplateID", imageFields, textFields, colorFields)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, templateResult, result)
}

func TestCanvaClient_UploadImageAssets(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient, testTokenBufferSecs)

	// Mock the responses for the UploadImageAssets API call
	mockClient.WillReturnBody("POST", assetUploadsEndpoint, `{
		"job": {
			"id": "1234",
			"status": "success",
			"asset": {
				"id": "5678"
			}
		}
	}`)

	mockClient.WillReturnBody("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("GET", "http://image1.jpg", `image1`)
	mockClient.WillReturnBody("GET", "http://image2.jpg", `image2`)

	images := []string{"http://image1.jpg", "http://image2.jpg"}
	imageIDs, err := canvaClient.UploadImageAssets(images)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, []string{"5678", "5678"}, imageIDs)
}

func TestCanvaClient_UploadColorAssets(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient, testTokenBufferSecs)

	// Mock the response for color asset uploads
	mockClient.WillReturnBody("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("POST", assetUploadsEndpoint, `{"job": {"id": "1234"}}`)
	mockClient.WillReturnBody("GET", assetUploadsEndpoint+"/1234", `{"job": {"status": "success", "asset": {"id": "colorID123"}}}`)

	colors := []string{"#FFFFFF", "#000000"}
	colorIDs, err := canvaClient.UploadColorAssets(colors)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, []string{"colorID123", "colorID123"}, colorIDs)
}

func TestCanvaClient_refreshAccessToken(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient, testTokenBufferSecs)

	// Mock the response for the token refresh
	mockClient.WillReturnBody("POST", tokenEndpoint+".*", `{"access_token": "newAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)

	token, err := canvaClient.refreshAccessToken()

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "newAccessToken", token)
}

func TestCanvaClient_sendAutofillRequest(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient, testTokenBufferSecs)

	// Mock access token
	mockClient.WillReturnBody("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)

	// Mock autofill request
	mockClient.WillReturnBody("POST", autofillEndpoint, `{"job": {"id": "1234"}}`)

	// Call sendAutofillRequest method
	data := map[string]interface{}{
		"brand_template_id": "testTemplateID",
		"data":              "testData",
	}

	resp, err := canvaClient.sendAutofillRequest(data)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCanvaClient_decodeUpdateTemplateJobResult_Success(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient, testTokenBufferSecs)

	// Mock the job status check
	mockClient.WillReturnBody("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("GET", autofillEndpoint+"/1234", `{"job": {"status": "success", "result": {"data": "testResult"}}}`)

	// Call decodeUpdateTemplateJobResult method
	result, err := canvaClient.decodeUpdateTemplateJobResult("1234")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCanvaClient_decodeUploadAssetResponse(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient, testTokenBufferSecs)

	// Mock the asset upload response
	mockClient.WillReturnBody("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("GET", assetUploadsEndpoint+"/1234", `{"job": {"status": "success", "asset": {"id": "assetID123"}}}`)

	// Call decodeUploadAssetResponse method
	resp, _ := mockClient.Get(assetUploadsEndpoint + "/1234")
	asset, err := canvaClient.decodeUploadAssetResponse(resp)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "assetID123", asset.ID)
}
