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
		canvaClient = NewClient("testClientID", "testClientSecret", "./test-canva-tokens.json", mockClient, testTokenBufferSecs)

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
	mockClient.WillReturnBodyRegex("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
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
	// given
	var (
		mockClient  = &http.MockHttpClient{}
		canvaClient = NewClient("testClientID", "testClientSecret", "./test-canva-tokens.json", mockClient, testTokenBufferSecs)

		images = []string{"http://image1.jpg", "http://image2.jpg"}
	)

	mockClient.WillReturnBody("POST", assetUploadsEndpoint, `{
		"job": {
			"id": "1234",
			"status": "success",
			"asset": {
				"id": "5678"
			}
		}
	}`)

	mockClient.WillReturnBodyRegex("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("GET", "http://image1.jpg", `image1`)
	mockClient.WillReturnBody("GET", "http://image2.jpg", `image2`)

	// when
	imageIDs, err := canvaClient.UploadImageAssets(images)

	// then
	assert.NoError(t, err)
	assert.Equal(t, []string{"5678", "5678"}, imageIDs)
}

func TestCanvaClient_UploadColorAssets(t *testing.T) {
	// given
	var (
		mockClient  = &http.MockHttpClient{}
		canvaClient = NewClient("testClientID", "testClientSecret", "./test-canva-tokens.json", mockClient, testTokenBufferSecs)

		colors = []string{"#FFFFFF", "#000000"}
	)

	mockClient.WillReturnBodyRegex("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("POST", assetUploadsEndpoint, `{"job": {"id": "1234"}}`)
	mockClient.WillReturnBody("GET", assetUploadsEndpoint+"/1234", `{"job": {"status": "success", "asset": {"id": "colorID123"}}}`)

	// when
	colorIDs, err := canvaClient.UploadColorAssets(colors)

	// then
	assert.NoError(t, err)
	assert.Equal(t, []string{"colorID123", "colorID123"}, colorIDs)
}

func TestCanvaClient_refreshAccessToken(t *testing.T) {
	// given
	var (
		mockClient  = &http.MockHttpClient{}
		canvaClient = NewClient("testClientID", "testClientSecret", "./test-canva-tokens.json", mockClient, testTokenBufferSecs)
	)

	mockClient.WillReturnBodyRegex("POST", tokenEndpoint+".*", `{"access_token": "newAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)

	// when
	token, err := canvaClient.refreshAccessToken()

	// then
	assert.NoError(t, err)
	assert.Equal(t, "newAccessToken", token)
}

func TestCanvaClient_sendAutofillRequest(t *testing.T) {
	// given
	var (
		mockClient  = &http.MockHttpClient{}
		canvaClient = NewClient("testClientID", "testClientSecret", "./test-canva-tokens.json", mockClient, testTokenBufferSecs)

		data = map[string]interface{}{
			"brand_template_id": "testTemplateID",
			"data":              "testData",
		}
	)

	mockClient.WillReturnBodyRegex("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("POST", autofillEndpoint, `{"job": {"id": "1234"}}`)

	// when
	resp, err := canvaClient.sendAutofillRequest(data)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCanvaClient_decodeUpdateTemplateJobResult_Success(t *testing.T) {
	// given
	var (
		mockClient  = &http.MockHttpClient{}
		canvaClient = NewClient("testClientID", "testClientSecret", "./test-canva-tokens.json", mockClient, testTokenBufferSecs)

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
	)

	mockClient.WillReturnBodyRegex("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
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
	result, err := canvaClient.decodeUpdateTemplateJobResult("1234")

	// then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, templateResult, result)
}

func TestCanvaClient_decodeUploadAssetResponse(t *testing.T) {
	// given
	var (
		mockClient  = &http.MockHttpClient{}
		canvaClient = NewClient("testClientID", "testClientSecret", "./test-canva-tokens.json", mockClient, testTokenBufferSecs)

		expectedAsset = &Asset{
			ID:        "asset_12345",
			Name:      "Winter Jacket",
			Tags:      []string{"clothing", "jacket", "winter"},
			CreatedAt: 1694640000,
			UpdatedAt: 1694643600,
			Thumbnail: Thumbnail{
				URL: "https://example.com/thumbnail.jpg",
			},
		}
	)

	mockClient.WillReturnBodyRegex("POST", tokenEndpoint+".*", `{"access_token": "validAccessToken", "expires_in": 0, "token_type": "Bearer", "refresh_token": "validRefreshToken"}`)
	mockClient.WillReturnBody("GET", assetUploadsEndpoint+"/1234", `{"job": {"status": "success", "asset": {
		"id": "asset_12345",
		"name": "Winter Jacket",
		"tags": ["clothing", "jacket", "winter"],
		"created_at": 1694640000,
		"updated_at": 1694643600,
		"thumbnail": {
			"url": "https://example.com/thumbnail.jpg"
		}
}}}`)

	// when
	resp, _ := mockClient.Get(assetUploadsEndpoint + "/1234")
	asset, err := canvaClient.decodeUploadAssetResponse(resp)

	// then
	assert.NoError(t, err)
	assert.Equal(t, "asset_12345", asset.ID)
	assert.Equal(t, asset, expectedAsset)
}
