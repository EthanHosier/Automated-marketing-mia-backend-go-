package researcher

import (
	"errors"
	"strings"
)

// MockResearcher is a mock implementation of the Researcher class for testing purposes.
type MockResearcher struct {
	sitemapResults                     map[string][]string
	businessSummaryResults             map[string]*BusinessSummary
	colorsFromUrlResults               map[string][]string
	pageContentsForResults             map[string]*PageContents
	pageBodyTextForResults             map[string]string
	socialMediaPostsForPlatformResults map[string]map[SocialMediaPlatform][]SocialMediaPost
	researchReportForResults           map[string]map[SocialMediaPlatform]string
	researchReportFromPostsResults     map[string]string
	googleAdsKeywordsDataResults       map[string][]GoogleAdsKeyword
	optimalKeywordsResults             map[string][2]string
	socialMediaPostsForResults         map[string][]SocialMediaPost

	// Use this to signal if an error should be returned
	sitemapError                     map[string]error
	businessSummaryError             map[string]error
	colorsFromUrlError               map[string]error
	pageContentsForError             map[string]error
	pageBodyTextForError             map[string]error
	socialMediaPostsForPlatformError map[string]map[SocialMediaPlatform]error
	researchReportForError           map[string]map[SocialMediaPlatform]error
	researchReportFromPostsError     map[string]error
	googleAdsKeywordsDataError       map[string]error
	optimalKeywordsError             map[string]error
	socialMediaPostsForError         map[string]error
}

// NewMockResearcher creates a new instance of MockResearcher.
func NewMockResearcher() *MockResearcher {
	return &MockResearcher{
		sitemapResults:                     make(map[string][]string),
		businessSummaryResults:             make(map[string]*BusinessSummary),
		colorsFromUrlResults:               make(map[string][]string),
		pageContentsForResults:             make(map[string]*PageContents),
		pageBodyTextForResults:             make(map[string]string),
		socialMediaPostsForPlatformResults: make(map[string]map[SocialMediaPlatform][]SocialMediaPost),
		researchReportForResults:           make(map[string]map[SocialMediaPlatform]string),
		researchReportFromPostsResults:     make(map[string]string),
		googleAdsKeywordsDataResults:       make(map[string][]GoogleAdsKeyword),
		optimalKeywordsResults:             make(map[string][2]string),
		socialMediaPostsForResults:         make(map[string][]SocialMediaPost),

		sitemapError:                     make(map[string]error),
		businessSummaryError:             make(map[string]error),
		colorsFromUrlError:               make(map[string]error),
		pageContentsForError:             make(map[string]error),
		pageBodyTextForError:             make(map[string]error),
		socialMediaPostsForPlatformError: make(map[string]map[SocialMediaPlatform]error),
		researchReportForError:           make(map[string]map[SocialMediaPlatform]error),
		researchReportFromPostsError:     make(map[string]error),
		googleAdsKeywordsDataError:       make(map[string]error),
		optimalKeywordsError:             make(map[string]error),
		socialMediaPostsForError:         make(map[string]error),
	}
}

// SitemapWillReturn sets the result for the Sitemap method.
func (m *MockResearcher) SitemapWillReturn(url string, result []string, err error) {
	m.sitemapResults[url] = result
	m.sitemapError[url] = err
}

// BusinessSummaryWillReturn sets the result for the BusinessSummary method.
func (m *MockResearcher) BusinessSummaryWillReturn(url string, result *BusinessSummary, err error) {
	m.businessSummaryResults[url] = result
	m.businessSummaryError[url] = err
}

// ColorsFromUrlWillReturn sets the result for the ColorsFromUrl method.
func (m *MockResearcher) ColorsFromUrlWillReturn(url string, result []string, err error) {
	m.colorsFromUrlResults[url] = result
	m.colorsFromUrlError[url] = err
}

// PageContentsForWillReturn sets the result for the PageContentsFor method.
func (m *MockResearcher) PageContentsForWillReturn(url string, result *PageContents, err error) {
	m.pageContentsForResults[url] = result
	m.pageContentsForError[url] = err
}

// PageBodyTextForWillReturn sets the result for the PageBodyTextFor method.
func (m *MockResearcher) PageBodyTextForWillReturn(url string, result string, err error) {
	m.pageBodyTextForResults[url] = result
	m.pageBodyTextForError[url] = err
}

// SocialMediaPostsForPlatformWillReturn sets the result for the SocialMediaPostsForPlatform method.
func (m *MockResearcher) SocialMediaPostsForPlatformWillReturn(keyword string, platform SocialMediaPlatform, result []SocialMediaPost, err error) {
	if _, ok := m.socialMediaPostsForPlatformResults[keyword]; !ok {
		m.socialMediaPostsForPlatformResults[keyword] = make(map[SocialMediaPlatform][]SocialMediaPost)
	}
	m.socialMediaPostsForPlatformResults[keyword][platform] = result
	if _, ok := m.socialMediaPostsForPlatformError[keyword]; !ok {
		m.socialMediaPostsForPlatformError[keyword] = make(map[SocialMediaPlatform]error)
	}
	m.socialMediaPostsForPlatformError[keyword][platform] = err
}

// ResearchReportForWillReturn sets the result for the ResearchReportFor method.
func (m *MockResearcher) ResearchReportForWillReturn(keyword string, platform SocialMediaPlatform, result string, err error) {
	if _, ok := m.researchReportForResults[keyword]; !ok {
		m.researchReportForResults[keyword] = make(map[SocialMediaPlatform]string)
	}
	m.researchReportForResults[keyword][platform] = result
	if _, ok := m.researchReportForError[keyword]; !ok {
		m.researchReportForError[keyword] = make(map[SocialMediaPlatform]error)
	}
	m.researchReportForError[keyword][platform] = err
}

// ResearchReportFromPostsWillReturn sets the result for the ResearchReportFromPosts method.
func (m *MockResearcher) ResearchReportFromPostsWillReturn(posts []SocialMediaPost, result string, err error) {
	keyword := ""
	if len(posts) > 0 {
		keyword = posts[0].Keyword
	}
	m.researchReportFromPostsResults[keyword] = result
	m.researchReportFromPostsError[keyword] = err
}

func (m *MockResearcher) SocialMediaPostsForWillReturn(keyword string, posts []SocialMediaPost, err error) {
	m.socialMediaPostsForResults[keyword] = posts
	m.socialMediaPostsForError[keyword] = err
}

// Implement the Researcher methods to use mocked results
func (m *MockResearcher) Sitemap(url string, timeout int) ([]string, error) {
	result, ok := m.sitemapResults[url]
	if !ok {
		return nil, errors.New("no result set for Sitemap")
	}
	err, _ := m.sitemapError[url]
	return result, err
}

func (m *MockResearcher) BusinessSummary(url string) ([]string, *BusinessSummary, []string, error) {
	result, ok := m.businessSummaryResults[url]
	if !ok {
		return nil, nil, nil, errors.New("no result set for BusinessSummary")
	}
	err, _ := m.businessSummaryError[url]
	return nil, result, nil, err
}

func (m *MockResearcher) ColorsFromUrl(url string) ([]string, error) {
	result, ok := m.colorsFromUrlResults[url]
	if !ok {
		return nil, errors.New("no result set for ColorsFromUrl")
	}
	err, _ := m.colorsFromUrlError[url]
	return result, err
}

func (m *MockResearcher) PageContentsFor(url string) (*PageContents, error) {
	result, ok := m.pageContentsForResults[url]
	if !ok {
		return nil, errors.New("no result set for PageContentsFor")
	}
	err, _ := m.pageContentsForError[url]
	return result, err
}

func (m *MockResearcher) PageBodyTextFor(url string) (string, error) {
	result, ok := m.pageBodyTextForResults[url]
	if !ok {
		return "", errors.New("no result set for PageBodyTextFor")
	}
	err, _ := m.pageBodyTextForError[url]
	return result, err
}

func (m *MockResearcher) SocialMediaPostsForPlatform(keyword string, platform SocialMediaPlatform) ([]SocialMediaPost, error) {
	result, ok := m.socialMediaPostsForPlatformResults[keyword][platform]
	if !ok {
		return nil, errors.New("no result set for SocialMediaPostsForPlatform")
	}
	err, _ := m.socialMediaPostsForPlatformError[keyword][platform]
	return result, err
}

func (m *MockResearcher) ResearchReportFor(keyword string, platform SocialMediaPlatform) (string, error) {
	result, ok := m.researchReportForResults[keyword][platform]
	if !ok {
		return "", errors.New("no result set for ResearchReportFor")
	}
	err, _ := m.researchReportForError[keyword][platform]
	return result, err
}

func (m *MockResearcher) ResearchReportFromPosts(posts []SocialMediaPost) (string, error) {
	keyword := ""
	if len(posts) > 0 {
		keyword = posts[0].Keyword
	}
	result, ok := m.researchReportFromPostsResults[keyword]
	if !ok {
		return "", errors.New("no result set for ResearchReportFromPosts")
	}
	err, _ := m.researchReportFromPostsError[keyword]
	return result, err
}

func (m *MockResearcher) GoogleAdsKeywordsDataWillReturn(keywords []string, result []GoogleAdsKeyword, err error) {
	key := strings.Join(keywords, ",")
	m.googleAdsKeywordsDataResults[key] = result
	m.googleAdsKeywordsDataError[key] = err
}

// OptimalKeywordsWillReturn sets the result for the OptimalKeywords method.
func (m *MockResearcher) OptimalKeywordsWillReturn(keywords []GoogleAdsKeyword, primaryKeyword, secondaryKeyword string, err error) {
	key := keywordsToString(keywords) // Helper function to convert slice of keywords to a unique string
	m.optimalKeywordsResults[key] = [2]string{primaryKeyword, secondaryKeyword}
	m.optimalKeywordsError[key] = err
}

// Implement GoogleAdsKeywordsData to use mocked results
func (m *MockResearcher) GoogleAdsKeywordsData(keywords []string) ([]GoogleAdsKeyword, error) {
	key := strings.Join(keywords, ",")
	result, ok := m.googleAdsKeywordsDataResults[key]
	if !ok {
		return nil, errors.New("no result set for GoogleAdsKeywordsData")
	}
	err := m.googleAdsKeywordsDataError[key]
	return result, err
}

// Implement OptimalKeywords to use mocked results
func (m *MockResearcher) OptimalKeywords(keywords []GoogleAdsKeyword) (string, string, error) {
	key := keywordsToString(keywords) // Helper function to generate unique key
	result, ok := m.optimalKeywordsResults[key]
	if !ok {
		return "", "", errors.New("no result set for OptimalKeywords")
	}
	err := m.optimalKeywordsError[key]
	return result[0], result[1], err
}

func (m *MockResearcher) SocialMediaPostsFor(keyword string) ([]SocialMediaPost, error) {
	result, ok := m.socialMediaPostsForResults[keyword]
	if !ok {
		return nil, errors.New("no result set for SocialMediaPostsFor")
	}
	err, _ := m.socialMediaPostsForError[keyword]
	return result, err
}

// Helper function to create a unique string from []GoogleAdsKeyword
func keywordsToString(keywords []GoogleAdsKeyword) string {
	var keyParts []string
	for _, kw := range keywords {
		keyParts = append(keyParts, kw.Keyword)
	}
	return strings.Join(keyParts, ",")
}
