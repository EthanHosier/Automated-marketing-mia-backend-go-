package openai

import (
	"context"
	"fmt"
)

type MockOpenaiClient struct {
	chatCompletionMocks  map[string]string
	imageCompletionMocks map[string]string
	embeddingsMocks      map[string][][]float32
	errorMocks           map[string]error
}

func (m *MockOpenaiClient) WillReturnChatCompletion(prompt string, model string, response string) {
	if m.chatCompletionMocks == nil {
		m.chatCompletionMocks = make(map[string]string)
	}
	key := fmt.Sprintf("%s:%s", prompt, model)
	m.chatCompletionMocks[key] = response
}

func (m *MockOpenaiClient) WillReturnImageCompletion(prompt string, images []string, model string, response string) {
	if m.imageCompletionMocks == nil {
		m.imageCompletionMocks = make(map[string]string)
	}
	key := fmt.Sprintf("%s:%v:%s", prompt, images, model)
	m.imageCompletionMocks[key] = response
}

func (m *MockOpenaiClient) WillReturnEmbeddings(urls []string, embeddings [][]float32) {
	if m.embeddingsMocks == nil {
		m.embeddingsMocks = make(map[string][][]float32)
	}
	key := fmt.Sprintf("%v", urls)
	m.embeddingsMocks[key] = embeddings
}

func (m *MockOpenaiClient) WillReturnError(err error) {
	if m.errorMocks == nil {
		m.errorMocks = make(map[string]error)
	}
	m.errorMocks["default"] = err
}

func (m *MockOpenaiClient) ChatCompletion(ctx context.Context, prompt string, model string) (string, error) {
	key := fmt.Sprintf("%s:%s", prompt, model)
	if err, ok := m.errorMocks["default"]; ok {
		return "", err
	}
	response, ok := m.chatCompletionMocks[key]
	if !ok {
		return "", fmt.Errorf("no chat completion mock found for prompt: %s, model: %s", prompt, model)
	}
	return response, nil
}

func (m *MockOpenaiClient) ImageCompletion(ctx context.Context, prompt string, images []string, model string) (string, error) {
	key := fmt.Sprintf("%s:%v:%s", prompt, images, model)
	if err, ok := m.errorMocks["default"]; ok {
		return "", err
	}
	response, ok := m.imageCompletionMocks[key]
	if !ok {
		return "", fmt.Errorf("no image completion mock found for prompt: %s, images: %v, model: %s", prompt, images, model)
	}
	return response, nil
}

func (m *MockOpenaiClient) Embeddings(urls []string) ([][]float32, error) {
	key := fmt.Sprintf("%v", urls)
	if err, ok := m.errorMocks["default"]; ok {
		return nil, err
	}
	embeddings, ok := m.embeddingsMocks[key]
	if !ok {
		return nil, fmt.Errorf("no embeddings mock found for URLs: %v", urls)
	}
	return embeddings, nil
}
