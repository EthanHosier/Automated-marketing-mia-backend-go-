package campaign_helper

import (
	"github.com/ethanhosier/mia-backend-go/canva"
	"github.com/ethanhosier/mia-backend-go/researcher"
)

type InitiFieldsResultsResult struct {
	textFields  []canva.TextField
	imageFields []canva.ImageField
	colorFields []canva.ColorField
}

type MockCampaignHelper struct {
	GetCandidatePageContentsForUserResults map[string][]researcher.PageContents
	GenerateThemesResults                  map[string][]CampaignTheme
	TemplatePlanResults                    map[string]*ExtractedTemplate
	InitFieldsResults                      map[string]InitiFieldsResultsResult

	GetCandidatePageContentsForUserErrs map[string]error
	GenerateThemesErrs                  map[string]error
	TemplatePlanErrs                    map[string]error
	InitFieldsErrs                      map[string]error
}

func NewMockCampaignHelper() *MockCampaignHelper {
	return &MockCampaignHelper{
		GetCandidatePageContentsForUserResults: map[string][]researcher.PageContents{},
		GenerateThemesResults:                  map[string][]CampaignTheme{},
		TemplatePlanResults:                    map[string]*ExtractedTemplate{},
		InitFieldsResults:                      map[string]InitiFieldsResultsResult{},

		GetCandidatePageContentsForUserErrs: map[string]error{},
		GenerateThemesErrs:                  map[string]error{},
		TemplatePlanErrs:                    map[string]error{},
		InitFieldsErrs:                      map[string]error{},
	}
}

func (m *MockCampaignHelper) GetCandidatePageContentsForUser(userID string, n int) ([]researcher.PageContents, error) {
	if err, ok := m.GetCandidatePageContentsForUserErrs[userID]; ok {
		return nil, err
	}
	return m.GetCandidatePageContentsForUserResults[userID], nil
}

func (m *MockCampaignHelper) GenerateThemes(pageContents []researcher.PageContents, businessSummary *researcher.BusinessSummary) ([]CampaignTheme, error) {
	if err, ok := m.GenerateThemesErrs[businessSummary.BusinessName]; ok {
		return nil, err
	}
	return m.GenerateThemesResults[businessSummary.BusinessName], nil
}

func (m *MockCampaignHelper) TemplatePlan(templatePrompt string) (*ExtractedTemplate, error) {
	if err, ok := m.TemplatePlanErrs[templatePrompt]; ok {
		return nil, err
	}
	return m.TemplatePlanResults[templatePrompt], nil
}

func (m *MockCampaignHelper) InitFields(template *ExtractedTemplate, campaignDetailsStr string, candidateImages []string) ([]canva.TextField, []canva.ImageField, []canva.ColorField, error) {
	if err, ok := m.InitFieldsErrs[campaignDetailsStr]; ok {
		return nil, nil, nil, err
	}
	result := m.InitFieldsResults[campaignDetailsStr]
	return result.textFields, result.imageFields, result.colorFields, nil
}

func (m *MockCampaignHelper) GetCandidatePageContentsForUserWillReturn(userID string, results []researcher.PageContents) {
	m.GetCandidatePageContentsForUserResults[userID] = results
}

func (m *MockCampaignHelper) GenerateThemesWillReturn(businessName string, results []CampaignTheme) {
	m.GenerateThemesResults[businessName] = results
}

func (m *MockCampaignHelper) TemplatePlanWillReturn(templatePrompt string, result *ExtractedTemplate) {
	m.TemplatePlanResults[templatePrompt] = result
}

func (m *MockCampaignHelper) InitFieldsWillReturn(campaignDetailsStr string, textFields []canva.TextField, imageFields []canva.ImageField, colorFields []canva.ColorField) {
	m.InitFieldsResults[campaignDetailsStr] = InitiFieldsResultsResult{
		textFields:  textFields,
		imageFields: imageFields,
		colorFields: colorFields,
	}
}
