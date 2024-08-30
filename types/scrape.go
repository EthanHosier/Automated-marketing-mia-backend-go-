package types

type BodyContentsScrapeResponse struct {
	Contents  WebsiteData `json:"contents"`
	ImageUrls []string    `json:"image_urls"`
	Url       string      `json:"url"`
}
