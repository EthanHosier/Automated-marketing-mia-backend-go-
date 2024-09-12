package services_test

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/services"
)

func TestPageScreenshot(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	url := "http://example.com"
	mockScreenshot := `{"screenshot": "mockedScreenshotData"}`
	mockClient.WillReturnBody(services.ScreenshotUrl+"?url="+url, mockScreenshot)

	result, err := sc.PageScreenshot(url)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "mockedScreenshotData"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSitemap(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	url := "http://example.com"
	timeout := 10
	mockSitemap := `["http://example.com/page1", "http://example.com/page2"]`
	mockClient.WillReturnBody(services.SitemapScraperUrl+"?url="+url+"&timeout="+fmt.Sprintf("%d", timeout), mockSitemap)

	result, err := sc.Sitemap(url, timeout)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"http://example.com/page1", "http://example.com/page2"}
	if len(result) != len(expected) {
		t.Errorf("expected %d URLs, got %d", len(expected), len(result))
	}
}

func TestScrapeSinglePageHtml(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	url := "http://example.com"
	mockHtml := "<html><body>mock page content</body></html>"
	mockClient.WillReturnBody(services.SinglePageHtmlScraperUrl+"?url="+url, mockHtml)

	result, err := sc.ScrapeSinglePageHtml(url)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != mockHtml {
		t.Errorf("expected %s, got %s", mockHtml, result)
	}
}

func TestGoogleAdsKeywordsData(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	keywords := []string{"keyword1", "keyword2"}
	mockResponse := `{
		"keywords": [
			{
				"keyword": "keyword1",
				"avg_monthly_searches": 1000,
				"competition_level": "high",
				"competition_index": 90,
				"low_top_of_page_bid": 10,
				"high_top_of_page_bid": 50
			},
			{
				"keyword": "keyword2",
				"avg_monthly_searches": 500,
				"competition_level": "medium",
				"competition_index": 60,
				"low_top_of_page_bid": 5,
				"high_top_of_page_bid": 30
			}
		]
	}`

	// Escape the keywords properly for the mock URL
	ks := "keyword1,keyword2"
	mockClient.WillReturnBody(services.GoogleAdsUrl+ks, mockResponse)

	result, err := sc.GoogleAdsKeywordsData(keywords)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}

	expectedFirstKeyword := services.GoogleAdsKeywordResponse{
		Keyword:            "keyword1",
		AvgMonthlySearches: 1000,
		CompetitionLevel:   "high",
		CompetitionIndex:   90,
		LowTopOfPageBid:    10,
		HighTopOfPageBid:   50,
	}

	expectedSecondKeyword := services.GoogleAdsKeywordResponse{
		Keyword:            "keyword2",
		AvgMonthlySearches: 500,
		CompetitionLevel:   "medium",
		CompetitionIndex:   60,
		LowTopOfPageBid:    5,
		HighTopOfPageBid:   30,
	}

	if result[0] != expectedFirstKeyword {
		t.Errorf("expected %+v, got %+v", expectedFirstKeyword, result[0])
	}

	if result[1] != expectedSecondKeyword {
		t.Errorf("expected %+v, got %+v", expectedSecondKeyword, result[1])
	}
}

func TestNumberOfSearchResultsFor(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	keyword := "test"
	mockResponse := `{"SearchResults": 123}`
	mockClient.WillReturnBody(services.SearchResultsUrl+url.QueryEscape(keyword), mockResponse)

	result, err := sc.NumberOfSearchResultsFor(keyword)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := 123
	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}
}

func TestScrapeSocialMediaFrom(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	keyword := "test"
	platform := "twitter"
	limit := 10
	mockResponse := `{
		"posts": [
			{
				"content": "post1 content",
				"hashtags": ["#example1", "#test1"],
				"url": "http://example.com/post1"
			},
			{
				"content": "post2 content",
				"hashtags": ["#example2", "#test2"],
				"url": "http://example.com/post2"
			}
		],
		"platform": "twitter"
	}`

	mockClient.WillReturnBody(
		services.SocialMediaFromKeywordScraperUrl+"?keyword="+url.QueryEscape(keyword)+"&platform="+platform+"&maxResults="+fmt.Sprintf("%d", limit),
		mockResponse,
	)

	result, err := sc.ScrapeSocialMediaFrom(keyword, platform, limit)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(result.Posts))
	}

	expectedFirstPost := services.SocialMediaFromKeywordPostResponse{
		Content:  "post1 content",
		Hashtags: []string{"#example1", "#test1"},
		Url:      "http://example.com/post1",
	}

	expectedSecondPost := services.SocialMediaFromKeywordPostResponse{
		Content:  "post2 content",
		Hashtags: []string{"#example2", "#test2"},
		Url:      "http://example.com/post2",
	}

	if result.Posts[0].Content != expectedFirstPost.Content {
		t.Errorf("expected content %s, got %s", expectedFirstPost.Content, result.Posts[0].Content)
	}
	if result.Posts[0].Url != expectedFirstPost.Url {
		t.Errorf("expected URL %s, got %s", expectedFirstPost.Url, result.Posts[0].Url)
	}
	if !equalStringSlices(result.Posts[0].Hashtags, expectedFirstPost.Hashtags) {
		t.Errorf("expected hashtags %+v, got %+v", expectedFirstPost.Hashtags, result.Posts[0].Hashtags)
	}

	if result.Posts[1].Content != expectedSecondPost.Content {
		t.Errorf("expected content %s, got %s", expectedSecondPost.Content, result.Posts[1].Content)
	}
	if result.Posts[1].Url != expectedSecondPost.Url {
		t.Errorf("expected URL %s, got %s", expectedSecondPost.Url, result.Posts[1].Url)
	}
	if !equalStringSlices(result.Posts[1].Hashtags, expectedSecondPost.Hashtags) {
		t.Errorf("expected hashtags %+v, got %+v", expectedSecondPost.Hashtags, result.Posts[1].Hashtags)
	}

	if result.Platform != platform {
		t.Errorf("expected platform %s, got %s", platform, result.Platform)
	}
}

func TestPageContentsScrape(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	testUrl := "http://example.com"
	mockResponse := `{
		"contents": {
			"title": "Example Title",
			"meta_description": "Example Meta Description",
			"headings": {
				"h1": ["Heading 1", "Heading 2"],
				"h2": ["Subheading 1"]
			},
			"keywords": "example, test",
			"links": ["http://link1.com", "http://link2.com"],
			"summary": "This is an example summary.",
			"categories": ["Category1", "Category2"]
		},
		"image_urls": ["http://image1.com", "http://image2.com"],
		"url": "http://example.com"
	}`

	// Mock the response for the given URL
	mockClient.WillReturnBody(services.SinglePageContentScraperUrl+testUrl, mockResponse)

	result, err := sc.PageContentsScrape(testUrl)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Expected result
	expectedResponse := &services.BodyContentsScrapeResponse{
		Contents: services.WebsiteData{
			Title:           "Example Title",
			MetaDescription: "Example Meta Description",
			Headings: map[string][]string{
				"h1": {"Heading 1", "Heading 2"},
				"h2": {"Subheading 1"},
			},
			Keywords:   "example, test",
			Links:      []string{"http://link1.com", "http://link2.com"},
			Summary:    "This is an example summary.",
			Categories: []string{"Category1", "Category2"},
		},
		ImageUrls: []string{"http://image1.com", "http://image2.com"},
		Url:       "http://example.com",
	}

	// Compare the fields individually
	if result.Url != expectedResponse.Url {
		t.Errorf("expected URL %s, got %s", expectedResponse.Url, result.Url)
	}

	if result.Contents.Title != expectedResponse.Contents.Title {
		t.Errorf("expected title %s, got %s", expectedResponse.Contents.Title, result.Contents.Title)
	}

	if result.Contents.MetaDescription != expectedResponse.Contents.MetaDescription {
		t.Errorf("expected meta description %s, got %s", expectedResponse.Contents.MetaDescription, result.Contents.MetaDescription)
	}

	if !equalStringSlices(result.Contents.Headings["h1"], expectedResponse.Contents.Headings["h1"]) {
		t.Errorf("expected h1 headings %v, got %v", expectedResponse.Contents.Headings["h1"], result.Contents.Headings["h1"])
	}

	if !equalStringSlices(result.Contents.Headings["h2"], expectedResponse.Contents.Headings["h2"]) {
		t.Errorf("expected h2 headings %v, got %v", expectedResponse.Contents.Headings["h2"], result.Contents.Headings["h2"])
	}

	if result.Contents.Keywords != expectedResponse.Contents.Keywords {
		t.Errorf("expected keywords %s, got %s", expectedResponse.Contents.Keywords, result.Contents.Keywords)
	}

	if !equalStringSlices(result.Contents.Links, expectedResponse.Contents.Links) {
		t.Errorf("expected links %v, got %v", expectedResponse.Contents.Links, result.Contents.Links)
	}

	if result.Contents.Summary != expectedResponse.Contents.Summary {
		t.Errorf("expected summary %s, got %s", expectedResponse.Contents.Summary, result.Contents.Summary)
	}

	if !equalStringSlices(result.Contents.Categories, expectedResponse.Contents.Categories) {
		t.Errorf("expected categories %v, got %v", expectedResponse.Contents.Categories, result.Contents.Categories)
	}

	if !equalStringSlices(result.ImageUrls, expectedResponse.ImageUrls) {
		t.Errorf("expected image URLs %v, got %v", expectedResponse.ImageUrls, result.ImageUrls)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestScrapeBusiness(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	testUrl := "http://example.com"
	mockResponse := `["http://page1.com", "http://page2.com", "http://page3.com"]`

	mockClient.WillReturnBody(services.BusinessScraperUrl+"?url="+testUrl+"&timeout="+fmt.Sprintf("%d", services.BusinessScrapeTimeout), mockResponse)

	result, err := sc.ScrapeBusiness(testUrl)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedScrapedPages := []string{
		"http://page1.com",
		"http://page2.com",
		"http://page3.com",
	}

	if len(result) != len(expectedScrapedPages) {
		t.Errorf("expected %d scraped pages, got %d", len(expectedScrapedPages), len(result))
	}

	for i, page := range result {
		if page != expectedScrapedPages[i] {
			t.Errorf("expected page %s at index %d, got %s", expectedScrapedPages[i], i, page)
		}
	}
}

func TestScrapeSinglePageBodyText(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	testUrl := "http://example.com"
	mockResponse := `{"content": "This is the body text of the page.", "url": "http://example.com"}`

	mockClient.WillReturnBody(services.SinglePageBodyTextScraperUrl+testUrl, mockResponse)

	content, err := sc.ScrapeSinglePageBodyText(testUrl)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedContent := "This is the body text of the page."

	if content != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, content)
	}
}

func TestPageScreenshot_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	url := "http://example.com"
	mockClient.WillReturnError(url, fmt.Errorf("mock error"))

	result, err := sc.PageScreenshot(url)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if result != "" {
		t.Errorf("expected empty result, got %s", result)
	}
}

func TestSitemap_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	url := "http://example.com"
	timeout := 10
	mockClient.WillReturnError(url, fmt.Errorf("mock error"))

	result, err := sc.Sitemap(url, timeout)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d URLs", len(result))
	}
}

func TestScrapeSinglePageHtml_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	url := "http://example.com"
	mockClient.WillReturnError(url, fmt.Errorf("mock error"))

	result, err := sc.ScrapeSinglePageHtml(url)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if result != "" {
		t.Errorf("expected empty result, got %s", result)
	}
}

func TestGoogleAdsKeywordsData_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	keywords := []string{"keyword1", "keyword2"}
	mockClient.WillReturnError("", fmt.Errorf("mock error"))

	result, err := sc.GoogleAdsKeywordsData(keywords)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d results", len(result))
	}
}

func TestNumberOfSearchResultsFor_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	keyword := "test"
	k := url.QueryEscape(keyword)

	expectedErr := fmt.Errorf("mock error")
	mockClient.WillReturnError(services.SearchResultsUrl+k, expectedErr)

	_, err := sc.NumberOfSearchResultsFor(keyword)

	if expectedErr != err {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

}

func TestScrapeSocialMediaFrom_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	keyword := "test"
	platform := "twitter"
	limit := 10

	url := services.SocialMediaFromKeywordScraperUrl + "?keyword=" + url.QueryEscape(keyword) + "&platform=" + platform + "&maxResults=" + strconv.Itoa(limit)
	expectedErr := fmt.Errorf("mock error")
	mockClient.WillReturnError(url, expectedErr)

	_, err := sc.ScrapeSocialMediaFrom(keyword, platform, limit)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestPageContentsScrape_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	testUrl := "http://example.com"
	expectedErr := fmt.Errorf("mock error")
	mockClient.WillReturnError(services.SinglePageContentScraperUrl+testUrl, expectedErr)

	_, err := sc.PageContentsScrape(testUrl)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestScrapeBusiness_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	testUrl := "http://example.com"
	finalUrl := services.BusinessScraperUrl + "?url=" + testUrl + "&timeout=" + fmt.Sprintf("%d", services.BusinessScrapeTimeout)
	expectedErr := fmt.Errorf("mock error")
	mockClient.WillReturnError(finalUrl, expectedErr)

	_, err := sc.ScrapeBusiness(testUrl)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestScrapeSinglePageBodyText_Error(t *testing.T) {
	mockClient := &http.MockHttpClient{}
	sc := services.NewServicesClient(mockClient)

	testUrl := "http://example.com"
	url := services.SinglePageBodyTextScraperUrl + testUrl
	expectedErr := fmt.Errorf("mock error")
	mockClient.WillReturnError(url, expectedErr)

	_, err := sc.ScrapeSinglePageBodyText(testUrl)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
