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

// Test successful post fetching for all platforms
func TestSocialMediaPostsFor_Success(t *testing.T) {
	mockResearcher := NewMockResearcher()

	instagramPosts := []SocialMediaPost{{Content: "Instagram Post 1"}, {Content: "Instagram Post 2"}}
	facebookPosts := []SocialMediaPost{{Content: "Facebook Post 1"}}
	linkedinPosts := []SocialMediaPost{{Content: "LinkedIn Post 1"}, {Content: "LinkedIn Post 2"}}

	posts := append(instagramPosts, facebookPosts...)
	posts = append(posts, linkedinPosts...)

	// Set mock results for platforms
	mockResearcher.SocialMediaPostsForWillReturn("fashion", posts, nil)

	// Test
	result, err := mockResearcher.SocialMediaPostsFor("fashion")

	assert.NoError(t, err, "expected no error but got one")
	assert.ElementsMatch(t, posts, result, "expected posts to match")
}

// Test fetching posts with one platform returning an error
func TestSocialMediaPostsFor_PartialError(t *testing.T) {
	mockResearcher := NewMockResearcher()

	posts := []SocialMediaPost{{Content: "Instagram Post 1"}, {Content: "Facebook Post 1"}}

	// Set mock results for platforms
	mockResearcher.SocialMediaPostsForWillReturn("fashion", posts, nil)

	// Test
	result, err := mockResearcher.SocialMediaPostsFor("fashion")

	assert.NoError(t, err, "expected no error but got one")
	assert.ElementsMatch(t, result, posts, "expected posts to match")
}

// Test when no results are set for one platform
func TestSocialMediaPostsFor_NoResultsForPlatform(t *testing.T) {
	mockResearcher := NewMockResearcher()
	instagramPosts := []SocialMediaPost{{Content: "Instagram Post 1"}}

	// Set mock results only for Instagram
	mockResearcher.SocialMediaPostsForWillReturn("fashion", instagramPosts, nil)

	// Test
	posts, err := mockResearcher.SocialMediaPostsFor("fashion")
	assert.NoError(t, err, "expected no error but got one")
	assert.ElementsMatch(t, instagramPosts, posts, "expected posts to match")
}

// Test if an error is returned when no results are set for any platform
func TestSocialMediaPostsFor_NoResultsError(t *testing.T) {
	mockResearcher := NewMockResearcher()

	// Test
	_, err := mockResearcher.SocialMediaPostsFor("fashion")

	assert.Error(t, err, "expected an error but got none")
	assert.Equal(t, "no result set for SocialMediaPostsFor", err.Error(), "unexpected error message")
}

func TestKeywordsToString(t *testing.T) {
	keywords := []GoogleAdsKeyword{{Keyword: "keyword1"}, {Keyword: "keyword2"}, {Keyword: "keyword3"}}
	expected := "keyword1,keyword2,keyword3"

	result := keywordsToString(keywords)

	assert.Equal(t, expected, result)
}

func TestGoogleAdsKeywordsData_Success(t *testing.T) {
	mockResearcher := NewMockResearcher()

	keywords := []string{"fashion", "style"}
	googleAdsKeywords := []GoogleAdsKeyword{
		{Keyword: "fashion", AvgMonthlySearches: 1000, CompetitionLevel: "High", CompetitionIndex: 5, LowTopOfPageBid: 10, HighTopOfPageBid: 20},
		{Keyword: "style", AvgMonthlySearches: 500, CompetitionLevel: "Medium", CompetitionIndex: 3, LowTopOfPageBid: 5, HighTopOfPageBid: 15},
	}

	mockResearcher.GoogleAdsKeywordsDataWillReturn(keywords, googleAdsKeywords, nil)

	// Test
	result, err := mockResearcher.GoogleAdsKeywordsData(keywords)

	assert.NoError(t, err, "expected no error but got one")
	assert.ElementsMatch(t, googleAdsKeywords, result, "expected Google Ads keywords data to match")
}

// Test GoogleAdsKeywordsData with an error
func TestGoogleAdsKeywordsData_Error(t *testing.T) {
	mockResearcher := NewMockResearcher()

	keywords := []string{"fashion", "style"}
	mockResearcher.GoogleAdsKeywordsDataWillReturn(keywords, nil, errors.New("failed to fetch Google Ads data"))

	// Test
	result, err := mockResearcher.GoogleAdsKeywordsData(keywords)

	assert.Error(t, err, "expected an error but got none")
	assert.Equal(t, "failed to fetch Google Ads data", err.Error(), "unexpected error message")
	assert.Nil(t, result, "expected result to be nil")
}

// Test OptimalKeywords with successful result
func TestOptimalKeywords_Success(t *testing.T) {
	mockResearcher := NewMockResearcher()

	keywords := []GoogleAdsKeyword{
		{Keyword: "fashion", AvgMonthlySearches: 1000, CompetitionLevel: "High", CompetitionIndex: 5, LowTopOfPageBid: 10, HighTopOfPageBid: 20},
		{Keyword: "style", AvgMonthlySearches: 500, CompetitionLevel: "Medium", CompetitionIndex: 3, LowTopOfPageBid: 5, HighTopOfPageBid: 15},
	}

	primaryKeyword := "fashion"
	secondaryKeyword := "style"

	mockResearcher.OptimalKeywordsWillReturn(keywords, primaryKeyword, secondaryKeyword, nil)

	// Test
	pKeyword, sKeyword, err := mockResearcher.OptimalKeywords(keywords)

	assert.NoError(t, err, "expected no error but got one")
	assert.Equal(t, primaryKeyword, pKeyword, "expected primary keyword to match")
	assert.Equal(t, secondaryKeyword, sKeyword, "expected secondary keyword to match")
}

// Test OptimalKeywords with an error
func TestOptimalKeywords_Error(t *testing.T) {
	mockResearcher := NewMockResearcher()

	keywords := []GoogleAdsKeyword{
		{Keyword: "fashion", AvgMonthlySearches: 1000, CompetitionLevel: "High", CompetitionIndex: 5, LowTopOfPageBid: 10, HighTopOfPageBid: 20},
	}

	mockResearcher.OptimalKeywordsWillReturn(keywords, "", "", errors.New("failed to determine optimal keywords"))

	// Test
	primaryKeyword, secondaryKeyword, err := mockResearcher.OptimalKeywords(keywords)

	assert.Error(t, err, "expected an error but got none")
	assert.Equal(t, "failed to determine optimal keywords", err.Error(), "unexpected error message")
	assert.Empty(t, primaryKeyword, "expected primary keyword to be empty")
	assert.Empty(t, secondaryKeyword, "expected secondary keyword to be empty")
}
