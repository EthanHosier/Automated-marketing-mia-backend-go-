package images

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockCaptionsFor(t *testing.T) {
	// given
	var (
		mockClient       = &MockImagesClient{}
		image            = "image1"
		expectedCaptions = []string{"caption1", "caption2"}
	)

	mockClient.WillReturnCaptionsFor(image, expectedCaptions)

	// when
	captions, err := mockClient.CaptionsFor(image)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedCaptions, captions)
}

func TestCaptionsForError(t *testing.T) {
	// given
	var (
		mockClient    = &MockImagesClient{}
		image         = "image1"
		expectedError = fmt.Errorf("error fetching captions")
	)

	mockClient.WillReturnCaptionsForError(image, expectedError)

	// when
	captions, err := mockClient.CaptionsFor(image)

	// then
	assert.Nil(t, captions)
	assert.Equal(t, expectedError, err)
}

func TestMockAiImageFrom(t *testing.T) {
	// given
	var (
		mockClient    = &MockImagesClient{}
		prompt        = "sunset"
		model         = StableImageCore
		expectedImage = []byte("generated-image")
	)

	mockClient.WillReturnAiImageFrom(prompt, model, expectedImage)

	// when
	image, err := mockClient.AiImageFrom(prompt, model)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedImage, image)
}

func TestAiImageFromError(t *testing.T) {
	var (
		mockClient    = &MockImagesClient{}
		prompt        = "sunset"
		model         = StableImageCore
		expectedError = fmt.Errorf("AI image generation failed")
	)

	mockClient.WillReturnAiImageFromError(prompt, model, expectedError)

	// when
	image, err := mockClient.AiImageFrom(prompt, model)

	// then
	assert.Nil(t, image)
	assert.Equal(t, expectedError, err)
}

func TestMockStockImageFrom(t *testing.T) {
	// given
	var (
		mockClient       = &MockImagesClient{}
		prompt           = "ocean"
		expectedImageURL = "http://example.com/ocean.jpg"
	)

	mockClient.WillReturnStockImageFrom(prompt, expectedImageURL)

	// when
	imageURL, err := mockClient.StockImageFrom(prompt)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedImageURL, imageURL)
}

func TestStockImageFromError(t *testing.T) {
	// given
	var (
		mockClient    = &MockImagesClient{}
		prompt        = "ocean"
		expectedError = fmt.Errorf("stock image not found")
	)

	mockClient.WillReturnStockImageFromError(prompt, expectedError)

	// when
	imageURL, err := mockClient.StockImageFrom(prompt)

	// then
	assert.Empty(t, imageURL)
	assert.Equal(t, expectedError, err)
}

func TestMockBestImageFor(t *testing.T) {
	// given
	var (
		mockClient      = &MockImagesClient{}
		desiredFeatures = []string{"feature1", "feature2"}
		prompt          = "car"
		expectedImage   = "http://example.com/best_car.jpg"
	)

	mockClient.WillReturnBestImageFor(desiredFeatures, prompt, expectedImage)

	// when
	image, err := mockClient.BestImageFor(nil, desiredFeatures, nil, "", prompt)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedImage, image)
}

func TestBestImageForError(t *testing.T) {
	// given
	var (
		mockClient      = &MockImagesClient{}
		desiredFeatures = []string{"feature1", "feature2"}
		prompt          = "car"
		expectedError   = fmt.Errorf("no matching best image found")
	)

	mockClient.WillReturnBestImageForError(desiredFeatures, prompt, expectedError)

	// when
	image, err := mockClient.BestImageFor(nil, desiredFeatures, nil, "", prompt)

	// then
	assert.Empty(t, image)
	assert.Equal(t, expectedError, err)
}
