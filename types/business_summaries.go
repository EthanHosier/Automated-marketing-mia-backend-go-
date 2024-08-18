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
}

type BusinessSummary struct {
	BusinessName    string `json:"businessName"`
	BusinessSummary string `json:"businessSummary"`
	BrandVoice      string `json:"brandVoice"`
	TargetRegion    string `json:"targetRegion"`
	TargetAudience  string `json:"targetAudience"`
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
