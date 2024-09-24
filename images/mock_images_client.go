package images

import (
	"context"
	"fmt"
)

type MockImagesClient struct {
	captionsForMocks          map[string][]string
	aiImageFromMocks          map[string][]byte
	stockImageFromMocks       map[string]string
	bestImageForMocks         map[string]string
	filterTooSmallImagesMocks map[string][]string

	captionsForError          map[string]error
	aiImageFromError          map[string]error
	stockImageFromError       map[string]error
	bestImageForError         map[string]error
	filterTooSmallImagesError map[string]error
}

func (m *MockImagesClient) WillReturnCaptionsFor(image string, captions []string) {
	if m.captionsForMocks == nil {
		m.captionsForMocks = make(map[string][]string)
	}
	m.captionsForMocks[image] = captions
}

func (m *MockImagesClient) WillReturnAiImageFrom(prompt string, model AiImageModel, result []byte) {
	if m.aiImageFromMocks == nil {
		m.aiImageFromMocks = make(map[string][]byte)
	}
	key := fmt.Sprintf("%s:%v", prompt, model)
	m.aiImageFromMocks[key] = result
}

func (m *MockImagesClient) WillReturnStockImageFrom(prompt string, result string) {
	if m.stockImageFromMocks == nil {
		m.stockImageFromMocks = make(map[string]string)
	}
	m.stockImageFromMocks[prompt] = result
}

func (m *MockImagesClient) WillReturnBestImageFor(ctxt context.Context, desiredFeatures []string, guaranteedImages []string, relevanceDescription string, prompt string, result string) {
	if m.bestImageForMocks == nil {
		m.bestImageForMocks = make(map[string]string)
	}
	key := fmt.Sprintf("%v:%s:%v:%v", desiredFeatures, prompt, guaranteedImages, relevanceDescription)
	m.bestImageForMocks[key] = result
}

func (m *MockImagesClient) WillReturnFilterTooSmallImages(images []string, result []string) {
	if m.filterTooSmallImagesMocks == nil {
		m.filterTooSmallImagesMocks = make(map[string][]string)
	}
	key := fmt.Sprintf("%v", images)
	m.filterTooSmallImagesMocks[key] = result
}

func (m *MockImagesClient) WillReturnCaptionsForError(image string, err error) {
	if m.captionsForError == nil {
		m.captionsForError = make(map[string]error)
	}
	m.captionsForError[image] = err
}

func (m *MockImagesClient) WillReturnAiImageFromError(prompt string, model AiImageModel, err error) {
	if m.aiImageFromError == nil {
		m.aiImageFromError = make(map[string]error)
	}
	key := fmt.Sprintf("%s:%v", prompt, model)
	m.aiImageFromError[key] = err
}

func (m *MockImagesClient) WillReturnStockImageFromError(prompt string, err error) {
	if m.stockImageFromError == nil {
		m.stockImageFromError = make(map[string]error)
	}
	m.stockImageFromError[prompt] = err
}

func (m *MockImagesClient) WillReturnBestImageForError(ctxt context.Context, desiredFeatures []string, guaranteedImages []string, relevanceDescription string, prompt string, err error) {
	if m.bestImageForError == nil {
		m.bestImageForError = make(map[string]error)
	}
	key := fmt.Sprintf("%v:%s:%v:%v", desiredFeatures, prompt, guaranteedImages, relevanceDescription)
	m.bestImageForError[key] = err
}

func (m *MockImagesClient) WillReturnFilterTooSmallImagesError(images []string, err error) {
	if m.filterTooSmallImagesError == nil {
		m.filterTooSmallImagesError = make(map[string]error)
	}
	key := fmt.Sprintf("%v", images)
	m.filterTooSmallImagesError[key] = err
}

func (m *MockImagesClient) CaptionsFor(image string) ([]string, error) {
	if m.captionsForError != nil {
		if err, ok := m.captionsForError[image]; ok {
			return nil, err
		}
	}

	if m.captionsForMocks != nil {
		if captions, ok := m.captionsForMocks[image]; ok {
			return captions, nil
		}
	}

	return nil, fmt.Errorf("no captions for image: %s", image)
}

func (m *MockImagesClient) AiImageFrom(prompt string, model AiImageModel) ([]byte, error) {
	key := fmt.Sprintf("%s:%v", prompt, model)
	if m.aiImageFromError != nil {
		if err, ok := m.aiImageFromError[key]; ok {
			return nil, err
		}
	}

	if m.aiImageFromMocks != nil {
		if result, ok := m.aiImageFromMocks[key]; ok {
			return result, nil
		}
	}

	return nil, fmt.Errorf("no ai image from prompt: %s, model: %v", prompt, model)
}

func (m *MockImagesClient) StockImageFrom(prompt string) (string, error) {
	if m.stockImageFromError != nil {
		if err, ok := m.stockImageFromError[prompt]; ok {
			return "", err
		}
	}

	if m.stockImageFromMocks != nil {
		if result, ok := m.stockImageFromMocks[prompt]; ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("no stock image from prompt: %s", prompt)
}

func (m *MockImagesClient) BestImageFor(ctxt context.Context, desiredFeatures []string, guaranteedImages []string, relevanceDescription string, prompt string) (string, error) {
	key := fmt.Sprintf("%v:%s:%v:%v", desiredFeatures, prompt, guaranteedImages, relevanceDescription)
	if m.bestImageForError != nil {
		if err, ok := m.bestImageForError[key]; ok {
			return "", err
		}
	}

	if m.bestImageForMocks != nil {
		if result, ok := m.bestImageForMocks[key]; ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("no best image for desired features: %v, prompt: %s, guaranteedImages: %v, relevanceDescription: %s", desiredFeatures, prompt, guaranteedImages, relevanceDescription)
}

func (m *MockImagesClient) CaptionsForAll(images []string) ([][]string, error) {
	var allCaptions [][]string
	for _, image := range images {
		captions, err := m.CaptionsFor(image)
		if err != nil {
			return nil, err
		}
		allCaptions = append(allCaptions, captions)
	}
	return allCaptions, nil
}

func (m *MockImagesClient) FilterTooSmallImages(images []string) ([]string, error) {
	key := fmt.Sprintf("%v", images)
	if m.filterTooSmallImagesError != nil {
		if err, ok := m.filterTooSmallImagesError[key]; ok {
			return nil, err
		}
	}

	if m.filterTooSmallImagesMocks != nil {
		if result, ok := m.filterTooSmallImagesMocks[key]; ok {
			return result, nil
		}
	}

	return nil, fmt.Errorf("no filtered images for images: %v", images)
}
