package handlers

import (
	"errors"
	"net/url"
)

type BusinessSummariesRequest struct {
	Url string `json:"url"`
}

func validateBusinessSummariesRequest(req BusinessSummariesRequest) error {
	if req.Url == "" {
		return errors.New("url is required")
	}

	_, err := url.ParseRequestURI(req.Url)
	if err != nil {
		return errors.New("invalid url format")
	}

	return nil
}
