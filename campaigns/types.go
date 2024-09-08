package campaigns

type CampaignTheme struct {
	Theme                         string `json:"theme"`
	Url                           string `json:"url"`
	SelectedUrl                   string `json:"selectedUrl"`
	ImageCanvaTemplateDescription string `json:"imageCanvaTemplateDescription"`
	PrimaryKeyword                string `json:"primaryKeyword"`
	SecondaryKeyword              string `json:"secondaryKeyword"`
}
