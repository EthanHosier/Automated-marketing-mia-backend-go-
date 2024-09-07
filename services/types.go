package services

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

type ScreenshotScraperResponse struct {
	ScreenshotBase64 string `json:"screenshot"`
}

type SinglePageBodyTextScraperResponse struct {
	Content string `json:"content"`
	Url     string `json:"url"`
}

type GoogleAdsKeyword struct {
	Keyword            string `json:"keyword"`
	AvgMonthlySearches int    `json:"avg_monthly_searches"`
	CompetitionLevel   string `json:"competition_level"`
	CompetitionIndex   int    `json:"competition_index"`
	LowTopOfPageBid    int    `json:"low_top_of_page_bid"`
	HighTopOfPageBid   int    `json:"high_top_of_page_bid"`
}

type GoogleAdsResponse struct {
	Keywords []GoogleAdsKeyword `json:"keywords"`
}

type SearchResultsResponse struct {
	SearchResults int `json:"searchResults"`
}
