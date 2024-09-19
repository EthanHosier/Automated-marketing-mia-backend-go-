package images

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
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

		desiredFeatures      = []string{"feature1", "feature2"}
		relevanceDescription = "some description"
		guaranteedImages     = []string{"gurl1", "gurl2"}

		vector1 = []float32{1, 2}
		vector2 = []float32{3, 4}

		feature1 = storage.ImageFeature{ID: "1", Feature: "a feature", FeatureEmbedding: vector1, UserId: "1", ImageUrl: "url1"}
		feature2 = storage.ImageFeature{ID: "2", Feature: "another feature", FeatureEmbedding: vector2, UserId: "2", ImageUrl: "url2"}

		allImages = []string{"url1", "url2"}
		imgPrompt = fmt.Sprintf(bestImagePrompt, relevanceDescription)
	)

	openaiClient.WillReturnEmbeddings(desiredFeatures, [][]float32{{1, 2}, {3, 4}})
	openaiClient.WillReturnImageCompletion(imgPrompt, append(guaranteedImages, allImages...), openai.GPT4o, "2")

	// when
	storage.StoreAll(store, feature1, feature2)
	resp, err := imagesClient.BestImageFor(context.Background(), desiredFeatures, guaranteedImages, relevanceDescription, "prompt")

	// then
	assert.NoError(t, err)
	assert.Equal(t, feature1.ImageUrl, resp)
}

func TestFilterImages(t *testing.T) {
	// given
	var (
		httpClient   = &http.MockHttpClient{}
		imagesClient = NewHttpImageClient(httpClient, nil, nil)

		smallImg = createImage(300, 500)
		largeImg = createImage(500, 500)

		urls = []string{"url1", "url2"}
	)

	httpClient.WillReturnBody("GET", "url1", string(smallImg))
	httpClient.WillReturnBody("GET", "url2", string(largeImg))

	// when
	filtered, err := imagesClient.FilterTooSmallImages(urls)

	// then
	assert.NoError(t, err)
	assert.Equal(t, []string{"url2"}, filtered)
}

func createImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the image with a solid color (for simplicity)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255} // red color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}

	// Encode the image to PNG format
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		panic(err) // Handle this more gracefully in a real application
	}

	return buf.Bytes()
}
