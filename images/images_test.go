package images

import (
	"testing"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestCaptionsFor(t *testing.T) {
	// given
	var (
		openaiClient = &openai.MockOpenaiClient{}
		imagesClient = NewHttpImageClient(nil, nil, openaiClient)

		image = "image1"
	)

	openaiClient.WillReturnImageCompletion(featuresPrompt, []string{image}, openai.GPT4o, `["caption1", "caption2"]`)

	// when
	captions, err := imagesClient.CaptionsFor(image)

	// then
	assert.NoError(t, err)
	assert.Equal(t, []string{"caption1", "caption2"}, captions)
}

func TestCaptionsForAll(t *testing.T) {
	// given
	var (
		openaiClient = &openai.MockOpenaiClient{}
		imagesClient = NewHttpImageClient(nil, nil, openaiClient)

		images = []string{"image1", "image2"}
	)

	openaiClient.WillReturnImageCompletion(featuresPrompt, []string{"image1"}, openai.GPT4o, `["caption1", "caption2"]`)
	openaiClient.WillReturnImageCompletion(featuresPrompt, []string{"image2"}, openai.GPT4o, `["caption3", "caption4"]`)

	// when
	captions, err := imagesClient.CaptionsForAll(images)

	// then
	assert.NoError(t, err)
	assert.Equal(t, [][]string{{"caption1", "caption2"}, {"caption3", "caption4"}}, captions)
}

func TestAiImageFrom(t *testing.T) {
	// given
	var (
		httpClient   = &http.MockHttpClient{}
		imagesClient = NewHttpImageClient(httpClient, nil, nil)

		model = StableImageCore
		url   = "https://api.stability.ai/v2beta/stable-image/generate/" + string(model)
	)

	httpClient.WillReturnBody("POST", url, "image")

	// when
	resp, err := imagesClient.AiImageFrom("prompt", model)

	// then
	assert.NoError(t, err)
	assert.Equal(t, []byte("image"), resp)
}

func TestStockImageFrom(t *testing.T) {
	// given
	var (
		imagesClient = NewHttpImageClient(nil, nil, nil)
	)

	// when
	resp, err := imagesClient.StockImageFrom("prompt")

	// then
	assert.NoError(t, err)
	assert.Equal(t, "", resp)
}

func TestBestImageFor(t *testing.T) {
	// given
	var (
		httpClient   = &http.MockHttpClient{}
		openaiClient = &openai.MockOpenaiClient{}
		store        = storage.NewInMemoryStorage()
		imagesClient = NewHttpImageClient(httpClient, store, openaiClient)

		desiredFeatures = []string{"feature1", "feature2"}
	)

	openaiClient.WillReturnEmbeddings(desiredFeatures, [][]float32{{1, 2}, {3, 4}})

	// when
	resp, err := imagesClient.BestImageFor(desiredFeatures, "prompt")

	// then
	assert.NoError(t, err)
	assert.Equal(t, "", resp)
}
