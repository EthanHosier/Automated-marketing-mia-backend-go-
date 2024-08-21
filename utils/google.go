package utils

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleClient struct {
	TokenSource oauth2.TokenSource
}

func NewGoogleClient() *GoogleClient {
	g := &GoogleClient{}
	g.initializeTokenSource()

	return g
}

func (g *GoogleClient) initializeTokenSource() error {
	data, err := os.ReadFile("mia-keyword-43b940a1c4ee.json")
	if err != nil {
		return fmt.Errorf("unable to read service account file: %w", err)
	}

	config, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/adwords")
	if err != nil {
		return fmt.Errorf("unable to create JWT config: %w", err)
	}

	g.TokenSource = config.TokenSource(context.Background())
	return nil
}
