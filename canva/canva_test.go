package canva

import (
	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCanvaClient_PopulateTemplate(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "./canva-tokens.json", mockClient)

	// Mock the response for the PopulateTemplate API call
	mockClient.WillReturnBody("POST", autofillEndpoint, `{"job": {"id": "1234"}}`)
	mockClient.WillReturnBody("GET", autofillEndpoint+"/1234", `{"job": {"status": "success"}}`)
	mockClient.WillReturnHeader("POST", tokenEndpoint+"?grant_type=refresh_token&refresh_token=", map[string]string{"Location": "https://redirect-url.com"})

	// Create test inputs
	imageFields := []ImageField{}
	textFields := []TextField{}
	colorFields := []ColorField{}

	// Call the method
	result, err := canvaClient.PopulateTemplate("testTemplateID", imageFields, textFields, colorFields)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCanvaClient_UploadImageAssets(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "/path/to/tokens.json", mockClient)

	// Mock the responses for the UploadImageAssets API call
	mockClient.WillReturnBody("POST", assetUploadsEndpoint, `{"job": {"id": "1234"}}`)
	mockClient.WillReturnBody("GET", assetUploadsEndpoint+"/1234", `{"job": {"status": "success", "asset": {"id": "assetID123"}}}`)

	images := []string{"image1.jpg", "image2.jpg"}
	imageIDs, err := canvaClient.UploadImageAssets(images)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, []string{"assetID123", "assetID123"}, imageIDs)
}

func TestCanvaClient_UploadColorAssets(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "/path/to/tokens.json", mockClient)

	// Mock the response for color asset uploads
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
	canvaClient := NewClient("testClientID", "testClientSecret", "/path/to/tokens.json", mockClient)

	// Mock the response for the token refresh
	mockClient.WillReturnBody("POST", tokenEndpoint, `{"access_token": "newAccessToken", "expires_in": 3600}`)

	token, err := canvaClient.refreshAccessToken()

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "newAccessToken", token)
}

func TestCanvaClient_sendAutofillRequest(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "/path/to/tokens.json", mockClient)

	// Mock access token
	mockClient.WillReturnBody("POST", tokenEndpoint, `{"access_token": "validAccessToken", "expires_in": 3600}`)

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
	canvaClient := NewClient("testClientID", "testClientSecret", "/path/to/tokens.json", mockClient)

	// Mock the job status check
	mockClient.WillReturnBody("GET", autofillEndpoint+"/1234", `{"job": {"status": "success", "result": {"data": "testResult"}}}`)

	// Call decodeUpdateTemplateJobResult method
	result, err := canvaClient.decodeUpdateTemplateJobResult("1234")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCanvaClient_decodeUploadAssetResponse(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	canvaClient := NewClient("testClientID", "testClientSecret", "/path/to/tokens.json", mockClient)

	// Mock the asset upload response
	mockClient.WillReturnBody("GET", assetUploadsEndpoint+"/1234", `{"job": {"status": "success", "asset": {"id": "assetID123"}}}`)

	// Call decodeUploadAssetResponse method
	resp, _ := mockClient.Get(assetUploadsEndpoint + "/1234")
	asset, err := canvaClient.decodeUploadAssetResponse(resp)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "assetID123", asset.ID)
}
