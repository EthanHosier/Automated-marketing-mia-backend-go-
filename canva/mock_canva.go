package canva

import (
	"fmt"
)

type MockCanvaClient struct {
	populateTemplateMocks  map[string]*UpdateTemplateResult
	uploadImageAssetsMocks map[string][]string
	uploadColorAssetsMocks map[string][]string
	populateTemplateError  error
	uploadImageAssetsError error
	uploadColorAssetsError error
}

func (m *MockCanvaClient) WillReturnPopulateTemplate(ID string, imageFields []ImageField, textFields []TextField, colorFields []ColorField, result *UpdateTemplateResult) {
	if m.populateTemplateMocks == nil {
		m.populateTemplateMocks = make(map[string]*UpdateTemplateResult)
	}
	key := fmt.Sprintf("%s:%v:%v:%v", ID, imageFields, textFields, colorFields)
	m.populateTemplateMocks[key] = result
}

func (m *MockCanvaClient) WillReturnUploadImageAssets(images []string, result []string) {
	if m.uploadImageAssetsMocks == nil {
		m.uploadImageAssetsMocks = make(map[string][]string)
	}
	key := fmt.Sprintf("%v", images)
	m.uploadImageAssetsMocks[key] = result
}

func (m *MockCanvaClient) WillReturnUploadColorAssets(colors []string, result []string) {
	if m.uploadColorAssetsMocks == nil {
		m.uploadColorAssetsMocks = make(map[string][]string)
	}
	key := fmt.Sprintf("%v", colors)
	m.uploadColorAssetsMocks[key] = result
}

func (m *MockCanvaClient) WillReturnPopulateTemplateError(err error) {
	m.populateTemplateError = err
}

func (m *MockCanvaClient) WillReturnUploadImageAssetsError(err error) {
	m.uploadImageAssetsError = err
}

func (m *MockCanvaClient) WillReturnUploadColorAssetsError(err error) {
	m.uploadColorAssetsError = err
}

func (m *MockCanvaClient) PopulateTemplate(ID string, imageFields []ImageField, textFields []TextField, colorFields []ColorField) (*UpdateTemplateResult, error) {
	key := fmt.Sprintf("%s:%v:%v:%v", ID, imageFields, textFields, colorFields)
	if m.populateTemplateError != nil {
		return nil, m.populateTemplateError
	}
	result, ok := m.populateTemplateMocks[key]
	if !ok {
		return nil, fmt.Errorf("no populate template mock found for ID: %s", ID)
	}
	return result, nil
}

func (m *MockCanvaClient) UploadImageAssets(images []string) ([]string, error) {
	key := fmt.Sprintf("%v", images)
	if m.uploadImageAssetsError != nil {
		return nil, m.uploadImageAssetsError
	}
	result, ok := m.uploadImageAssetsMocks[key]
	if !ok {
		return nil, fmt.Errorf("no upload image assets mock found for images: %v", images)
	}
	return result, nil
}

func (m *MockCanvaClient) UploadColorAssets(colors []string) ([]string, error) {
	key := fmt.Sprintf("%v", colors)
	if m.uploadColorAssetsError != nil {
		return nil, m.uploadColorAssetsError
	}
	result, ok := m.uploadColorAssetsMocks[key]
	if !ok {
		return nil, fmt.Errorf("no upload color assets mock found for colors: %v", colors)
	}
	return result, nil
}
