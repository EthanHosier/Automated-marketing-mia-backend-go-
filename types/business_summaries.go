package types

import (
	"errors"
	"net/url"
)

type BusinessSummariesRequest struct {
	Url string `json:"url"`
}

type BusinessSummariesResponse struct {
	ScreenshotBase64 string `json:"screenshot"`
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
