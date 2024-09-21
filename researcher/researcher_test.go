package researcher

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSitemap(t *testing.T) {
	// given
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		expectedUrls = []string{"http://example.com/page1", "http://example.com/page2"}
	)

	mockHttpClient.WillReturnBody("GET", services.SitemapScraperUrl+"?url=http://example.com&timeout=15", `["http://example.com/page1", "http://example.com/page2"]`)

	// when
	urls, err := researcher.Sitemap("http://example.com", 15)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, expectedUrls, urls)
}

func TestBusinessSummary(t *testing.T) {
	// given
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		expectedUrls = []string{"http://example.com/page1", "http://example.com/page2"}

		pageContents1 = &PageContents{
			TextContents: services.WebsiteData{
				Title:           "Page Title 1",
				MetaDescription: "Page Description 1",
				Headings: map[string][]string{
					"H1": {"Heading 1"},
				},
				Keywords:   "keyword1, keyword2",
				Links:      []string{"http://example.com/page1/link1", "http://example.com/page1/link2"},
				Summary:    "Page summary",
				Categories: []string{"Category 1", "Category 2"},
			},
			ImageUrls: []string{"http://example.com/page1/image.jpg"},
			Url:       "http://example.com/page1",
		}

		pageContents2 = &PageContents{
			TextContents: services.WebsiteData{
				Title:           "Page Title 2",
				MetaDescription: "Page Description 2",
				Headings: map[string][]string{
					"H1": {"Heading 2"},
				},
				Keywords:   "keyword3, keyword4",
				Links:      []string{"http://example.com/page2/link1", "http://example.com/page2/link2"},
				Summary:    "Page summary",
				Categories: []string{"Category 3", "Category 4"},
			},
			ImageUrls: []string{"http://example.com/page2/image.jpg"},
			Url:       "http://example.com/page2",
		}

		pageContents = []string{pageContents2.TextContents.String(), pageContents1.TextContents.String()}
		jsonData, _  = json.Marshal(pageContents)

		prompt = openai.BusinessSummaryPrompt + string(jsonData)
	)

	jsonData1, err := json.Marshal(pageContents1)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	jsonData2, err := json.Marshal(pageContents2)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	mockHttpClient.WillReturnBody("GET", services.SitemapScraperUrl+"?url=http://example.com&timeout=15", `["http://example.com/page1", "http://example.com/page2"]`)
	mockHttpClient.WillReturnBody("GET", services.SinglePageContentScraperUrl+"http://example.com/page1", string(jsonData1))
	mockHttpClient.WillReturnBody("GET", services.SinglePageContentScraperUrl+"http://example.com/page2", string(jsonData2))
	mockHttpClient.WillReturnBody("GET", services.ScreenshotUrl+"?url=http://example.com", `{"screenshot": "mockedBase64Image"}`)

	mockOpenaiClient.WillReturnChatCompletion(prompt, openai.GPT4o, `{"summary": "Business summary"}`)
	mockOpenaiClient.WillReturnImageCompletion(openai.ColorThemesPrompt, []string{"mockedBase64Image"}, openai.GPT4o, `["#FFFFFF", "#000000"]`)

	// when
	urls, summary, imageUrls, err := researcher.BusinessSummary("http://example.com")

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, expectedUrls, urls)
	assert.NotNil(t, summary)
	assert.ElementsMatch(t, imageUrls, []string{"http://example.com/page2/image.jpg", "http://example.com/page1/image.jpg"})
}

func TestColorsFromUrl(t *testing.T) {
	// given
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		screenshotBase64 = "mockedBase64Image"
		expectedColors   = []string{"#FFFFFF", "#000000"}
	)

	mockHttpClient.WillReturnBody("GET", services.ScreenshotUrl+"?url=http://example.com", `{"screenshot": "mockedBase64Image"}`)

	mockOpenaiClient.WillReturnImageCompletion(openai.ColorThemesPrompt, []string{screenshotBase64}, openai.GPT4o, `["#FFFFFF", "#000000"]`)

	// when
	colors, err := researcher.ColorsFromUrl("http://example.com")

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, expectedColors, colors)
}

func TestPageContentsFor(t *testing.T) {
	// given
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		expectedContent = &PageContents{
			TextContents: services.WebsiteData{
				Title: "Page Title",
			},
			ImageUrls: []string{"http://example.com/image.jpg"},
			Url:       "http://example.com",
		}
	)

	mockHttpClient.WillReturnBody("GET", services.SinglePageContentScraperUrl+"http://example.com",
		`{"contents":{"Title": "Page Title"}, "image_urls": ["http://example.com/image.jpg"], "url": "http://example.com"}`)

	// when
	contents, err := researcher.PageContentsFor("http://example.com")

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedContent, contents)
}

func TestSocialMediaPostsForPlatform(t *testing.T) {
	// given
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		expectedPosts = []SocialMediaPost{
			{
				Content:  "Post content",
				Hashtags: []string{"#example"},
				Url:      "http://example.com/post",
				Platform: Instagram,
				Keyword:  "keyword",
			},
		}
	)

	mockHttpClient.WillReturnBody("GET", services.SocialMediaFromKeywordScraperUrl+"?keyword=keyword&platform=instagram&maxResults=5", `{"posts": [{"content": "Post content", "hashtags": ["#example"], "url": "http://example.com/post"}]}`)

	// when
	posts, err := researcher.SocialMediaPostsForPlatform("keyword", Instagram)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, expectedPosts, posts)
}

func TestResearchReportFor(t *testing.T) {
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		keyword       = "keyword"
		expectedPosts = []SocialMediaPost{
			{
				Content:  "Post content",
				Hashtags: []string{"#example"},
				Url:      "http://example.com/post",
				Platform: Instagram,
				Keyword:  "keyword",
			},
		}
		prompt         = fmt.Sprintf(openai.ResearchReportPrompt, keyword, expectedPosts)
		expectedReport = "Research report"
	)

	mockHttpClient.WillReturnBody("GET", services.SocialMediaFromKeywordScraperUrl+"?keyword=keyword&platform=instagram&maxResults=5", `{"posts": [{"content": "Post content", "hashtags": ["#example"], "url": "http://example.com/post"}]}`)

	mockOpenaiClient.WillReturnChatCompletion(prompt, openai.GPT4oMini, expectedReport)

	// when
	report, err := researcher.ResearchReportFor("keyword", Instagram)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedReport, report)
}

func TestResearchReportFromPosts(t *testing.T) {
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		posts = []SocialMediaPost{
			{
				Content:  "Post content",
				Hashtags: []string{"#example"},
				Url:      "http://example.com/post",
				Platform: Instagram,
				Keyword:  "keyword",
			},
		}
		expectedReport = "Research report"
		prompt         = fmt.Sprintf(openai.ResearchReportPrompt, "keyword", posts)
	)

	mockOpenaiClient.WillReturnChatCompletion(prompt, openai.GPT4oMini, expectedReport)

	// when
	report, err := researcher.ResearchReportFromPosts(posts)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedReport, report)
}

func TestSocialMediaPostsFor(t *testing.T) {
	// given
	var (
		mockHttpClient   = &http.MockHttpClient{}
		mockOpenaiClient = &openai.MockOpenaiClient{}

		servicesClient = services.NewServicesClient(mockHttpClient)
		researcher     = New(servicesClient, mockOpenaiClient)

		keyword       = "keyword"
		expectedPosts = map[SocialMediaPlatform][]SocialMediaPost{
			Instagram: {
				{
					Content:  "Instagram Post",
					Hashtags: []string{"#example"},
					Url:      "http://example.com/instagram-post",
					Platform: Instagram,
					Keyword:  keyword,
				},
			},
			Facebook: {
				{
					Content:  "Facebook Post",
					Hashtags: []string{"#example"},
					Url:      "http://example.com/facebook-post",
					Platform: Facebook,
					Keyword:  keyword,
				},
			},
			LinkedIn: {
				{
					Content:  "LinkedIn Post",
					Hashtags: []string{"#example"},
					Url:      "http://example.com/linkedin-post",
					Platform: LinkedIn,
					Keyword:  keyword,
				},
			},
			Google: {
				{
					Content:  "Google Post",
					Hashtags: []string{"#example"},
					Url:      "http://example.com/google-post",
					Platform: Google,
					Keyword:  keyword,
				},
			},
			News: {
				{
					Content:  "News Post",
					Hashtags: []string{"#example"},
					Url:      "http://example.com/news-post",
					Platform: News,
					Keyword:  keyword,
				},
			},
		}
	)

	// Mocking the responses for each platform
	for platform, posts := range expectedPosts {
		postsJson, _ := json.Marshal(posts)
		resp := `{"posts": ` + string(postsJson) + `, "platform": "` + string(platform) + `"}`

		mockHttpClient.WillReturnBody("GET", services.SocialMediaFromKeywordScraperUrl+"?keyword="+keyword+"&platform="+string(platform)+"&maxResults=5", resp)
	}

	// when
	actualPosts, err := researcher.SocialMediaPostsFor(keyword)

	// then
	require.NoError(t, err)

	var expectedPostsFlat []SocialMediaPost
	for _, posts := range expectedPosts {
		expectedPostsFlat = append(expectedPostsFlat, posts...)
	}

	assert.ElementsMatch(t, expectedPostsFlat, actualPosts)
}

func TestEmbeddingsFor(t *testing.T) {
	// given
	var (
		mockOpenaiClient = &openai.MockOpenaiClient{}
		researcher       = New(nil, mockOpenaiClient)

		urls               = []string{"http://example.com/page1", "http://example.com/page2"}
		expectedEmbeddings = [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}}
	)

	mockOpenaiClient.WillReturnEmbeddings(urls, expectedEmbeddings)

	// when
	embeddings, err := researcher.EmbeddingsFor(urls)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedEmbeddings, embeddings)
}
