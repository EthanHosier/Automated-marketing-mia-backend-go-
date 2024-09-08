package researcher

import "github.com/ethanhosier/mia-backend-go/services"

type BusinessSummary struct {
	BusinessName    string   `json:"businessName"`
	BusinessSummary string   `json:"businessSummary"`
	BrandVoice      string   `json:"brandVoice"`
	TargetRegion    string   `json:"targetRegion"`
	TargetAudience  string   `json:"targetAudience"`
	Colors          []string `json:"colors"`
}

type SitemapUrl string

type PageContents struct {
	TextContents services.WebsiteData `json:"contents"`
	ImageUrls    []string             `json:"image_urls"`
	Url          string               `json:"url"`
}

type GoogleAdsKeyword struct {
	Keyword            string `json:"keyword"`
	AvgMonthlySearches int    `json:"avg_monthly_searches"`
	CompetitionLevel   string `json:"competition_level"`
	CompetitionIndex   int    `json:"competition_index"`
	LowTopOfPageBid    int    `json:"low_top_of_page_bid"`
	HighTopOfPageBid   int    `json:"high_top_of_page_bid"`
}
