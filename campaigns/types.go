package campaigns

type CampaignTheme struct {
	Theme                         string   `json:"theme"`
	Keywords                      []string `json:"keywords"`
	Url                           string   `json:"url"`
	SelectedUrl                   string   `json:"selectedUrl"`
	ImageCanvaTemplateDescription string   `json:"imageCanvaTemplateDescription"`
}
