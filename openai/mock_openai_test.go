package openai_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockOpenaiClient_ChatCompletion(t *testing.T) {
	mockClient := &openai.MockOpenaiClient{}
	mockClient.WillReturnChatCompletion("test prompt", "test model", "test response")

	response, err := mockClient.ChatCompletion(context.Background(), "test prompt", "test model")
	require.NoError(t, err)
	assert.Equal(t, "test response", response)
}

func TestMockOpenaiClient_ChatCompletion_Error(t *testing.T) {
	mockClient := &openai.MockOpenaiClient{}
	expectedError := fmt.Errorf("chat completion error")
	mockClient.WillReturnError(expectedError)

	response, err := mockClient.ChatCompletion(context.Background(), "test prompt", "test model")
	assert.Error(t, err)
	assert.Equal(t, err, expectedError)
	assert.Equal(t, "", response)
}

func TestMockOpenaiClient_ImageCompletion(t *testing.T) {
	mockClient := &openai.MockOpenaiClient{}
	mockClient.WillReturnImageCompletion("test prompt", []string{"image1", "image2"}, "test model", "test response")

	response, err := mockClient.ImageCompletion(context.Background(), "test prompt", []string{"image1", "image2"}, "test model")
	require.NoError(t, err)
	assert.Equal(t, "test response", response)
}

func TestMockOpenaiClient_ImageCompletion_Error(t *testing.T) {
	mockClient := &openai.MockOpenaiClient{}
	expectedError := fmt.Errorf("image completion error")
	mockClient.WillReturnError(expectedError)

	response, err := mockClient.ImageCompletion(context.Background(), "test prompt", []string{"image1"}, "test model")
	assert.Error(t, err)
	assert.Equal(t, err, expectedError)
	assert.Equal(t, "", response)
}

func TestMockOpenaiClient_Embeddings(t *testing.T) {
	mockClient := &openai.MockOpenaiClient{}
	expectedEmbeddings := [][]float32{{0.1, 0.2}, {0.3, 0.4}}
	mockClient.WillReturnEmbeddings([]string{"url1", "url2"}, expectedEmbeddings)

	embeddings, err := mockClient.Embeddings([]string{"url1", "url2"})
	require.NoError(t, err)
	assert.Equal(t, expectedEmbeddings, embeddings)
}

func TestMockOpenaiClient_Embeddings_Error(t *testing.T) {
	mockClient := &openai.MockOpenaiClient{}
	expectedError := fmt.Errorf("embeddings error")
	mockClient.WillReturnError(fmt.Errorf("embeddings error"))

	embeddings, err := mockClient.Embeddings([]string{"url1"})
	assert.Error(t, err)
	assert.Equal(t, err, expectedError)
	assert.Nil(t, embeddings)
}

func TestMockOpenaiClient_NoMockFound(t *testing.T) {
	mockClient := &openai.MockOpenaiClient{}

	// Test ChatCompletion without setting a mock
	expectedError := fmt.Errorf("no chat completion mock found for prompt: unknown prompt, model: unknown model")
	response, err := mockClient.ChatCompletion(context.Background(), "unknown prompt", "unknown model")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", response)

	// Test ImageCompletion without setting a mock
	expectedError = fmt.Errorf("no image completion mock found for prompt: unknown prompt, images: [unknown image], model: unknown model")
	response, err = mockClient.ImageCompletion(context.Background(), "unknown prompt", []string{"unknown image"}, "unknown model")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", response)

	// Test Embeddings without setting a mock
	expectedError = fmt.Errorf("no embeddings mock found for URLs: [unknown url]")
	embeddings, err := mockClient.Embeddings([]string{"unknown url"})
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, embeddings)
}
