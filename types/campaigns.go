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
	Theme                         string   `json:"theme"`
	Keywords                      []string `json:"keywords"`
	Url                           string   `json:"url"`
	SelectedUrl                   string   `json:"selectedUrl"`
	ImageCanvaTemplateDescription string   `json:"imageCanvaTemplateDescription"`
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

type TemplateFields struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Label         string `json:"label"`
	Comment       string `json:"comment"`
	Page          string `json:"page"`
	MaxCharacters int    `json:"maxCharacters"`
}

type ColorField struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Comment string `json:"comment"`
}

type NearestTemplateResponse struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Platforms   []string         `json:"platforms"`
	ExportType  string           `json:"export_type"`
	Description string           `json:"description"`
	Fields      []TemplateFields `json:"fields"`
	ColorFields []ColorField     `json:"colors"`
	Similarity  float32          `json:"similarity"`
}

type SocialMediaFromKeywordPostResponse struct {
	Content  string   `json:"content"`
	Hashtags []string `json:"hashtags"`
	Url      string   `json:"url"`
}

type SummarisedPost struct {
	Content  string   `json:"content"`
	Url      string   `json:"url"`
	Hashtags []string `json:"hashtags"`
}

type SocialMediaFromKeywordResponse struct {
	Posts    []SocialMediaFromKeywordPostResponse `json:"posts"`
	Platform string                               `json:"platform"`
}

type PlatformResearchReport struct {
	Platform string                               `json:"platform"`
	Posts    []SocialMediaFromKeywordPostResponse `json:"summarisedPosts"`
}

type ResearchReportData struct {
	PlatformResearchReports []PlatformResearchReport `json:"platformResearchReports"`
}

type FieldType string

const (
	TextType  FieldType = "text"
	ImageType FieldType = "image"
)

type PopulatedField struct {
	Name  string    `json:"name"`
	Value string    `json:"value"`
	Type  FieldType `json:"type"`
}

type PopulatedColorField struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type PopulatedTemplate struct {
	Platform    string                `json:"platform"`
	Fields      []PopulatedField      `json:"fields"`
	ColorFields []PopulatedColorField `json:"colors"`
	Caption     string                `json:"caption"`
}

type TemplateAndCaption struct {
	TemplateResult UpdateTemplateResult `json:"template"`
	Caption        string               `json:"caption"`
}

type GeneratedCampaign struct {
	Posts            []TemplateAndCaption `json:"posts"`
	PrimaryKeyword   string               `json:"primaryKeyword"`
	SecondaryKeyword string               `json:"secondaryKeyword"`
	Theme            string               `json:"theme"`
}
