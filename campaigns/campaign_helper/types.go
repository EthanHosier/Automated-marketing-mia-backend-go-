package campaign_helper

type FieldType string

const (
	TextType  FieldType = "text"
	ImageType FieldType = "image"
)

type ExtractedTemplate struct {
	Platform    string                `json:"platform"`
	Fields      []PopulatedField      `json:"fields"`
	ColorFields []PopulatedColorField `json:"colors"`
	Caption     string                `json:"caption"`
}

type PopulatedField struct {
	Name  string    `json:"name"`
	Value string    `json:"value"`
	Type  FieldType `json:"type"`
}

type PopulatedColorField struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CampaignTheme struct {
	Theme                         string `json:"theme"`
	Url                           string `json:"url"`
	SelectedUrl                   string `json:"selectedUrl"`
	ImageCanvaTemplateDescription string `json:"imageCanvaTemplateDescription"`
	PrimaryKeyword                string `json:"primaryKeyword"`
	SecondaryKeyword              string `json:"secondaryKeyword"`
}

type themeWithSuggestedKeywords struct {
	Theme                         string   `json:"theme"`
	Keywords                      []string `json:"keywords"`
	Url                           string   `json:"url"`
	SelectedUrl                   string   `json:"selectedUrl"`
	ImageCanvaTemplateDescription string   `json:"imageCanvaTemplateDescription"`
}
