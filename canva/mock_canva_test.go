package canva

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPopulateTemplate(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Setup mock response
	expectedResult := &UpdateTemplateResult{
		Type:   "MockType",
		Design: Design{
			// Populate with expected Design fields
		},
	}
	mockClient.WillReturnPopulateTemplate("templateID", []ImageField{}, []TextField{}, []ColorField{}, expectedResult)

	// Test
	result, err := mockClient.PopulateTemplate("templateID", []ImageField{}, []TextField{}, []ColorField{})
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestUploadImageAssets(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Setup mock response
	images := []string{"image1.png", "image2.png"}
	expectedResult := []string{"asset1", "asset2"}
	mockClient.WillReturnUploadImageAssets(images, expectedResult)

	// Test
	result, err := mockClient.UploadImageAssets(images)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestUploadColorAssets(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Setup mock response
	colors := []string{"#FF5733", "#33FF57"}
	expectedResult := []string{"colorAsset1", "colorAsset2"}
	mockClient.WillReturnUploadColorAssets(colors, expectedResult)

	// Test
	result, err := mockClient.UploadColorAssets(colors)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestPopulateTemplateWithError(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Setup specific error
	mockClient.WillReturnPopulateTemplateError(fmt.Errorf("populate template error"))

	// Test
	result, err := mockClient.PopulateTemplate("templateID", []ImageField{}, []TextField{}, []ColorField{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "populate template error", err.Error())
}

func TestUploadImageAssetsWithError(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Setup specific error
	mockClient.WillReturnUploadImageAssetsError(fmt.Errorf("upload image assets error"))

	// Test
	result, err := mockClient.UploadImageAssets([]string{"image1.png", "image2.png"})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "upload image assets error", err.Error())
}

func TestUploadColorAssetsWithError(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Setup specific error
	mockClient.WillReturnUploadColorAssetsError(fmt.Errorf("upload color assets error"))

	// Test
	result, err := mockClient.UploadColorAssets([]string{"#FF5733", "#33FF57"})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "upload color assets error", err.Error())
}

func TestPopulateTemplateNoMockFound(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Test with no mock set up
	result, err := mockClient.PopulateTemplate("unknownID", []ImageField{}, []TextField{}, []ColorField{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "no populate template mock found for ID: unknownID", err.Error())
}

func TestUploadImageAssetsNoMockFound(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Test with no mock set up
	result, err := mockClient.UploadImageAssets([]string{"unknownImage.png"})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "no upload image assets mock found for images: [unknownImage.png]", err.Error())
}

func TestUploadColorAssetsNoMockFound(t *testing.T) {
	mockClient := &MockCanvaClient{}

	// Test with no mock set up
	result, err := mockClient.UploadColorAssets([]string{"#000000"})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "no upload color assets mock found for colors: [#000000]", err.Error())
}
