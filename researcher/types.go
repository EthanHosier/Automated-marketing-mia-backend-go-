package researcher

type BusinessSummary struct {
	BusinessName    string   `json:"businessName"`
	BusinessSummary string   `json:"businessSummary"`
	BrandVoice      string   `json:"brandVoice"`
	TargetRegion    string   `json:"targetRegion"`
	TargetAudience  string   `json:"targetAudience"`
	Colors          []string `json:"colors"`
}

type SitemapUrl string
