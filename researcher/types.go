package researcher

import "github.com/ethanhosier/mia-backend-go/services"

type BusinessSummary struct {
	ID              string   `json:"id"`
	BusinessName    string   `json:"businessName"`
	BusinessSummary string   `json:"businessSummary"`
	BrandVoice      string   `json:"brandVoice"`
	TargetRegion    string   `json:"targetRegion"`
	TargetAudience  string   `json:"targetAudience"`
	Colors          []string `json:"colors"`
}

type SitemapUrl struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

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

type SocialMediaPost struct {
	Platform SocialMediaPlatform `json:"platform"`
	Content  string              `json:"content"`
	Hashtags []string            `json:"hashtags"`
	Url      string              `json:"url"`
	Keyword  string              `json:"keyword"`
}

type SocialMediaPlatform string

const (
	Instagram SocialMediaPlatform = "instagram"
	Facebook  SocialMediaPlatform = "facebook"
	LinkedIn  SocialMediaPlatform = "linkedIn"
	Google    SocialMediaPlatform = "google"
	News      SocialMediaPlatform = "news"
)

var SocialMediaPlatforms = []SocialMediaPlatform{Instagram, Facebook, LinkedIn, Google, News}
