package types

type ScreenshotScraperResponse struct {
	ScreenshotBase64 string `json:"screenshot"`
}

type SinglePageBodyTextScraperResponse struct {
	Content string `json:"content"`
	Url     string `json:"url"`
}
