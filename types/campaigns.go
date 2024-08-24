package types

type GenerateCampaignsRequest struct {
	Instructions           string `json:"instructions"`
	TargetAudienceLocation string `json:"targetAudienceLocation"`
	Backlink               string `json:"backlink"`
	ImageBase64            string `json:"imageBase64"`
}

type GenerateCampaignsResponse struct {
	OptimalKeywords []OptimalKeyword `json:"optimalkeywords"`
}

type ThemeData struct {
	Theme       string   `json:"theme"`
	Keywords    []string `json:"keywords"`
	SelectedUrl string   `json:"selectedUrl"`
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

type Vector []float32

type AdsKeywordsResult struct {
	Theme       string
	SelectedUrl string
	AdsData     []GoogleAdsKeyword
}

type OptimalKeyword struct {
	Keyword     string `json:"keyword"`
	SelectedUrl string `json:"selectedUrl"`
}

type SearchResultsResponse struct {
	SearchResults int `json:"searchResults"`
}
