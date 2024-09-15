package campaign_helper

import (
	"errors"
	"testing"

	"github.com/ethanhosier/mia-backend-go/canva"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestMockGetCandidatePageContentsForUser(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedResults := []researcher.PageContents{{}} // Adjust according to the actual structure
	mock.GetCandidatePageContentsForUserWillReturn("user1", expectedResults)

	results, err := mock.GetCandidatePageContentsForUser("user1", 10)
	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
}

func TestMockGenerateThemes(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedResults := []CampaignTheme{{}} // Adjust according to the actual structure
	mock.GenerateThemesWillReturn("business1", expectedResults)
	businessSummary := &researcher.BusinessSummary{BusinessName: "business1"}

	results, err := mock.GenerateThemes([]researcher.PageContents{}, businessSummary)
	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
}

func TestMockTemplatePlan(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedResult := &ExtractedTemplate{} // Adjust according to the actual structure
	mock.TemplatePlanWillReturn("prompt1", expectedResult)

	result, err := mock.TemplatePlan("prompt1", &storage.Template{})
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestMockInitFields(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedTextFields := []canva.TextField{
		{Name: "text1", Text: "Sample text"},
	} // Adjust according to the actual structure
	expectedImageFields := []canva.ImageField{
		{Name: "image1", AssetId: "asset123"},
	} // Adjust according to the actual structure
	expectedColorFields := []canva.ColorField{
		{Name: "color1", ColorAssetId: "color123"},
	} // Adjust according to the actual structure
	mock.InitFieldsWillReturn("details1", expectedTextFields, expectedImageFields, expectedColorFields)

	textFields, imageFields, colorFields, err := mock.InitFields(&ExtractedTemplate{}, "details1", []string{})
	assert.NoError(t, err)
	assert.Equal(t, expectedTextFields, textFields)
	assert.Equal(t, expectedImageFields, imageFields)
	assert.Equal(t, expectedColorFields, colorFields)
}

func TestMockGetCandidatePageContentsForUserError(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedErr := errors.New("error fetching page contents")
	mock.GetCandidatePageContentsForUserErrs["user1"] = expectedErr

	results, err := mock.GetCandidatePageContentsForUser("user1", 10)
	assert.Nil(t, results)
	assert.Equal(t, expectedErr, err)
}

func TestMockGenerateThemesError(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedErr := errors.New("error generating themes")
	mock.GenerateThemesErrs["business1"] = expectedErr
	businessSummary := &researcher.BusinessSummary{BusinessName: "business1"}

	results, err := mock.GenerateThemes([]researcher.PageContents{}, businessSummary)
	assert.Nil(t, results)
	assert.Equal(t, expectedErr, err)
}

func TestMockTemplatePlanError(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedErr := errors.New("error planning template")
	mock.TemplatePlanErrs["prompt1"] = expectedErr

	result, err := mock.TemplatePlan("prompt1", &storage.Template{})
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestMockInitFieldsError(t *testing.T) {
	mock := NewMockCampaignHelper()
	expectedErr := errors.New("error initializing fields")
	mock.InitFieldsErrs["details1"] = expectedErr

	textFields, imageFields, colorFields, err := mock.InitFields(&ExtractedTemplate{}, "details1", []string{})
	assert.Nil(t, textFields)
	assert.Nil(t, imageFields)
	assert.Nil(t, colorFields)
	assert.Equal(t, expectedErr, err)
}
