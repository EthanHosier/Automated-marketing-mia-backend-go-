package images

import "fmt"

type MockImagesClient struct {
	captionsForMocks    map[string][]string
	aiImageFromMocks    map[string][]byte
	stockImageFromMocks map[string]string
	bestImageForMocks   map[string]string

	captionsForError    map[string]error
	aiImageFromError    map[string]error
	stockImageFromError map[string]error
	bestImageForError   map[string]error
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

func (m *MockImagesClient) WillReturnBestImageFor(desiredFeatures []string, prompt string, result string) {
	if m.bestImageForMocks == nil {
		m.bestImageForMocks = make(map[string]string)
	}
	key := fmt.Sprintf("%v:%s", desiredFeatures, prompt)
	m.bestImageForMocks[key] = result
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

func (m *MockImagesClient) WillReturnBestImageForError(desiredFeatures []string, prompt string, err error) {
	if m.bestImageForError == nil {
		m.bestImageForError = make(map[string]error)
	}
	key := fmt.Sprintf("%v:%s", desiredFeatures, prompt)
	m.bestImageForError[key] = err
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

func (m *MockImagesClient) BestImageFor(desiredFeatures []string, prompt string) (string, error) {
	key := fmt.Sprintf("%v:%s", desiredFeatures, prompt)
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

	return "", fmt.Errorf("no best image for desired features: %v, prompt: %s", desiredFeatures, prompt)
}
