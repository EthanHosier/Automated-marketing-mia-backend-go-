package types

type GenerateCampaignsRequest struct {
	Instructions           string `json:"instructions"`
	TargetAudienceLocation string `json:"targetAudienceLocation"`
	Backlink               string `json:"backlink"`
	ImageBase64            string `json:"imageBase64"`
}

type GenerateCampaignsResponse struct {
	Themes []ThemeData `json:"themes"`
}

type ThemeData struct {
	Theme       string   `json:"theme"`
	Keywords    []string `json:"keywords"`
	SelectedUrl string   `json:"selectedUrl"`
}
