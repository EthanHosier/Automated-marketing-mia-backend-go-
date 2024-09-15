package researcher

import (
	"errors"
	"testing"

	"github.com/ethanhosier/mia-backend-go/services"
	"github.com/stretchr/testify/assert"
)

func TestSitemapWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := []string{"url1", "url2"}
	expectedError := error(nil)

	mock.SitemapWillReturn("test-url", expectedResult, expectedError)

	result, err := mock.Sitemap("test-url", 15)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestBusinessSummaryWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := &BusinessSummary{
		BusinessName:    "Test Business",
		BusinessSummary: "A summary of the business",
		BrandVoice:      "Friendly",
		TargetRegion:    "Global",
		TargetAudience:  "Consumers",
		Colors:          []string{"Red", "Blue"},
	}
	expectedError := error(nil)

	mock.BusinessSummaryWillReturn("test-url", expectedResult, expectedError)

	_, result, _, err := mock.BusinessSummary("test-url")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestColorsFromUrlWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := []string{"Red", "Green"}
	expectedError := error(nil)

	mock.ColorsFromUrlWillReturn("test-url", expectedResult, expectedError)

	result, err := mock.ColorsFromUrl("test-url")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestPageContentsForWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := &PageContents{
		TextContents: services.WebsiteData{
			Title:           "Test Title",
			MetaDescription: "Test Description",
		},
		ImageUrls: []string{"image1.jpg", "image2.jpg"},
		Url:       "test-url",
	}
	expectedError := error(nil)

	mock.PageContentsForWillReturn("test-url", expectedResult, expectedError)

	result, err := mock.PageContentsFor("test-url")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestPageBodyTextForWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := "Test body text"
	expectedError := error(nil)

	mock.PageBodyTextForWillReturn("test-url", expectedResult, expectedError)

	result, err := mock.PageBodyTextFor("test-url")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestSocialMediaPostsForPlatformWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := []SocialMediaPost{
		{Platform: Instagram, Content: "Post content", Hashtags: []string{"#hashtag"}, Url: "post-url", Keyword: "keyword"},
	}
	expectedError := error(nil)

	mock.SocialMediaPostsForPlatformWillReturn("keyword", Instagram, expectedResult, expectedError)

	result, err := mock.SocialMediaPostsForPlatform("keyword", Instagram)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestResearchReportForWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := "Research report content"
	expectedError := error(nil)

	mock.ResearchReportForWillReturn("keyword", Instagram, expectedResult, expectedError)

	result, err := mock.ResearchReportFor("keyword", Instagram)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestResearchReportFromPostsWillReturn(t *testing.T) {
	mock := NewMockResearcher()
	expectedResult := "Research report content from posts"
	expectedError := error(nil)

	mock.ResearchReportFromPostsWillReturn([]SocialMediaPost{
		{Keyword: "keyword"},
	}, expectedResult, expectedError)

	result, err := mock.ResearchReportFromPosts([]SocialMediaPost{
		{Keyword: "keyword"},
	})
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestSitemapWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to get sitemap")

	mock.SitemapWillReturn("test-url", nil, expectedError)

	result, err := mock.Sitemap("test-url", 15)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestBusinessSummaryWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to get business summary")

	mock.BusinessSummaryWillReturn("test-url", nil, expectedError)

	_, result, _, err := mock.BusinessSummary("test-url")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestColorsFromUrlWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to get colors")

	mock.ColorsFromUrlWillReturn("test-url", nil, expectedError)

	result, err := mock.ColorsFromUrl("test-url")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestPageContentsForWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to get page contents")

	mock.PageContentsForWillReturn("test-url", nil, expectedError)

	result, err := mock.PageContentsFor("test-url")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestPageBodyTextForWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to get page body text")

	mock.PageBodyTextForWillReturn("test-url", "", expectedError)

	result, err := mock.PageBodyTextFor("test-url")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
}

func TestSocialMediaPostsForPlatformWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to get social media posts")

	mock.SocialMediaPostsForPlatformWillReturn("keyword", Instagram, nil, expectedError)

	result, err := mock.SocialMediaPostsForPlatform("keyword", Instagram)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestResearchReportForWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to generate research report")

	mock.ResearchReportForWillReturn("keyword", Instagram, "", expectedError)

	result, err := mock.ResearchReportFor("keyword", Instagram)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
}

func TestResearchReportFromPostsWillReturnError(t *testing.T) {
	mock := NewMockResearcher()
	expectedError := errors.New("failed to generate research report from posts")

	mock.ResearchReportFromPostsWillReturn([]SocialMediaPost{}, "", expectedError)

	result, err := mock.ResearchReportFromPosts([]SocialMediaPost{})
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
}
