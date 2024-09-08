package services

import (
	"fmt"
	"strings"
)

type BodyContentsScrapeResponse struct {
	Contents  WebsiteData `json:"contents"`
	ImageUrls []string    `json:"image_urls"`
	Url       string      `json:"url"`
}

type WebsiteData struct {
	Title           string              `json:"title"`
	MetaDescription string              `json:"meta_description"`
	Headings        map[string][]string `json:"headings"`
	Keywords        string              `json:"keywords"`
	Links           []string            `json:"links"`
	Summary         string              `json:"summary"`
	Categories      []string            `json:"categories"`
}

func (w WebsiteData) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title: %s\n", w.Title))
	sb.WriteString(fmt.Sprintf("Meta Description: %s\n", w.MetaDescription))

	sb.WriteString("Headings:\n")
	for key, values := range w.Headings {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", key, strings.Join(values, ", ")))
	}

	sb.WriteString(fmt.Sprintf("Keywords: %s\n", w.Keywords))
	sb.WriteString(fmt.Sprintf("Links: %s\n", strings.Join(w.Links, ", ")))
	sb.WriteString(fmt.Sprintf("Summary: %s\n", w.Summary))
	sb.WriteString(fmt.Sprintf("Categories: %s\n", strings.Join(w.Categories, ", ")))

	return sb.String()
}

type ScreenshotScraperResponse struct {
	ScreenshotBase64 string `json:"screenshot"`
}

type SinglePageBodyTextScraperResponse struct {
	Content string `json:"content"`
	Url     string `json:"url"`
}

type GoogleAdsKeywordResponse struct {
	Keyword            string `json:"keyword"`
	AvgMonthlySearches int    `json:"avg_monthly_searches"`
	CompetitionLevel   string `json:"competition_level"`
	CompetitionIndex   int    `json:"competition_index"`
	LowTopOfPageBid    int    `json:"low_top_of_page_bid"`
	HighTopOfPageBid   int    `json:"high_top_of_page_bid"`
}

type GoogleAdsResponse struct {
	Keywords []GoogleAdsKeywordResponse `json:"keywords"`
}

type SearchResultsResponse struct {
	SearchResults int `json:"searchResults"`
}
