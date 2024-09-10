package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	ScreenshotUrl                    = "https://p21i96yy3f.execute-api.eu-west-2.amazonaws.com/screenshot-scraper"
	SitemapScraperUrl                = "https://f5z5c23uk7.execute-api.eu-west-2.amazonaws.com/sitemap-scraper"
	BusinessScraperUrl               = "https://qkns6w88sc.execute-api.eu-west-2.amazonaws.com/business-scraper"
	GoogleAdsUrl                     = "https://629o94qd23.execute-api.eu-west-2.amazonaws.com/google-ads?keywords="
	SearchResultsUrl                 = "https://4azcvme8md.execute-api.eu-west-2.amazonaws.com/search-results-scraper?keyword="
	SocialMediaFromKeywordScraperUrl = "https://rszmhstlr0.execute-api.eu-west-2.amazonaws.com/social-media-from-keyword-scraper"
	SinglePageBodyTextScraperUrl     = "https://33lz2xpok8.execute-api.eu-west-2.amazonaws.com/single-page-body-text-scraper?url="
	SinglePageHtmlScraperUrl         = "https://5s8ecjywfi.execute-api.eu-west-2.amazonaws.com/single-page-html-scraper"
	SinglePageContentScraperUrl      = "https://1ap5f1w55b.execute-api.eu-west-2.amazonaws.com/single-page-content-scraper?url="

	businessScrapeTimeout = 15
)

type ServicesClient struct {
	httpClient *http.Client
}

func NewServicesClient(httpClient *http.Client) *ServicesClient {
	return &ServicesClient{
		httpClient: httpClient,
	}
}

func (sc *ServicesClient) PageScreenshot(url string) (string, error) {
	resp, err := sc.httpClient.Get(ScreenshotUrl + "?url=" + url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response ScreenshotScraperResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.ScreenshotBase64, nil
}

func (sc *ServicesClient) Sitemap(url string, timeout int) ([]string, error) {
	resp, err := sc.httpClient.Get(SitemapScraperUrl + "?url=" + url + "&timeout=" + fmt.Sprintf("%d", timeout))
	if err != nil {
		return []string{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var urls []string
	err = json.Unmarshal(body, &urls)
	if err != nil {
		return nil, err
	}

	// Filter out non-URL websites (e.g., ending in .xml, .pdf)
	var filteredUrls []string
	for _, u := range urls {
		if !strings.HasSuffix(u, ".xml") && !strings.HasSuffix(u, ".pdf") {
			filteredUrls = append(filteredUrls, u)
		}
	}

	return filteredUrls, nil
}

func (sc *ServicesClient) ScrapeSinglePageHtml(url string) (string, error) {
	resp, err := http.Get(SinglePageHtmlScraperUrl + "?url=" + url)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (sc *ServicesClient) ScrapeSinglePageBodyText(url string) (string, error) {
	resp, err := http.Get(SinglePageBodyTextScraperUrl + url)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response SinglePageBodyTextScraperResponse
	err = json.Unmarshal(body, &response)

	return response.Content, err
}

func (sc *ServicesClient) ScrapeBusiness(url string) ([]string, error) {
	resp, err := http.Get(BusinessScraperUrl + "?url=" + url + "&timeout=" + fmt.Sprintf("%d", businessScrapeTimeout))
	if err != nil {
		return []string{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var scrapedPages []string
	err = json.Unmarshal(body, &scrapedPages)

	return scrapedPages, err
}

func (sc *ServicesClient) PageContentsScrape(url string) (*BodyContentsScrapeResponse, error) {
	resp, err := sc.httpClient.Get(SinglePageContentScraperUrl + url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response BodyContentsScrapeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (sc *ServicesClient) GoogleAdsKeywordsData(keywords []string) ([]GoogleAdsKeywordResponse, error) {
	queryKeywords := []string{}

	for _, keyword := range keywords {
		queryKeywords = append(queryKeywords, url.QueryEscape(keyword))
	}

	keywordsStr := strings.Join(queryKeywords, ",")

	resp, err := sc.httpClient.Get(GoogleAdsUrl + keywordsStr)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response GoogleAdsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response.Keywords, nil
}

func (sc *ServicesClient) NumberOfSearchResultsFor(keyword string) (int, error) {
	k := url.QueryEscape(keyword)
	resp, err := sc.httpClient.Get(SearchResultsUrl + k)

	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	var response SearchResultsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return -1, err
	}

	return response.SearchResults, nil
}

func (sc *ServicesClient) ScrapeSocialMediaFrom(keyword string, platform string, limit int) (*SocialMediaFromKeywordResponse, error) {
	resp, err := sc.httpClient.Get(SocialMediaFromKeywordScraperUrl + "?keyword=" + url.QueryEscape(keyword) + "&platform=" + platform + "&maxResults=" + strconv.Itoa(limit))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response SocialMediaFromKeywordResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
