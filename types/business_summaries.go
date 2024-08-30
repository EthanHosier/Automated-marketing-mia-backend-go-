package types

import (
	"errors"
	"net/url"
)

type BusinessSummariesRequest struct {
	Url string `json:"url"`
}

type BusinessSummariesResponse struct {
	BusinessSummaries BusinessSummary `json:"businessSummaries"`
	ImageUrls         []string        `json:"imageUrls"`
}

type BusinessSummary struct {
	BusinessName    string `json:"businessName"`
	BusinessSummary string `json:"businessSummary"`
	BrandVoice      string `json:"brandVoice"`
	TargetRegion    string `json:"targetRegion"`
	TargetAudience  string `json:"targetAudience"`
}

type StoredBusinessSummary struct {
	ID              string `json:"id"`
	BusinessName    string `json:"businessName"`
	BusinessSummary string `json:"businessSummary"`
	BrandVoice      string `json:"brandVoice"`
	TargetRegion    string `json:"targetRegion"`
	TargetAudience  string `json:"targetAudience"`
}

type StoredSitemapUrl struct {
	ID     string `json:"id"`
	Url    string `json:"url"`
	Vector Vector `json:"url_embedding"`
}

type SitemapResponse struct {
	Urls []string `json:"urls"`
}

func ValidateBusinessSummariesRequest(req BusinessSummariesRequest) error {
	if req.Url == "" {
		return errors.New("url is required")
	}

	_, err := url.ParseRequestURI(req.Url)
	if err != nil {
		return errors.New("invalid url format")
	}

	return nil
}

type UrlHtmlPair struct {
	Url  string `json:"url"`
	Html string `json:"html"`
}
