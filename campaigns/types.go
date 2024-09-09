package campaigns

import "github.com/ethanhosier/mia-backend-go/researcher"

type FieldType string

const (
	TextType  FieldType = "text"
	ImageType FieldType = "image"
)

type CampaignTheme struct {
	Theme                         string `json:"theme"`
	Url                           string `json:"url"`
	SelectedUrl                   string `json:"selectedUrl"`
	ImageCanvaTemplateDescription string `json:"imageCanvaTemplateDescription"`
	PrimaryKeyword                string `json:"primaryKeyword"`
	SecondaryKeyword              string `json:"secondaryKeyword"`
}

type SocialMediaResearch struct {
	Platform string `json:"platform"`
	Posts    []researcher.SocialMediaPost
	Keyword  string `json:"keyword"`
}

// shouldnt be used anywhere else, but need to work with json unmarshalling
type PopulatedField struct {
	Name  string    `json:"name"`
	Value string    `json:"value"`
	Type  FieldType `json:"type"`
}

type PopulatedColorField struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type ExtractedTemplate struct {
	Platform    string                `json:"platform"`
	Fields      []PopulatedField      `json:"fields"`
	ColorFields []PopulatedColorField `json:"colors"`
	Caption     string                `json:"caption"`
}
