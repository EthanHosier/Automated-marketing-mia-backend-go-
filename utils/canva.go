package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethanhosier/mia-backend-go/types"
)

const (
	tokensFilePath = "canva-tokens.json"
)

var (
	mu sync.Mutex
)

func loadTokens() (*types.Token, error) {
	file, err := os.ReadFile(tokensFilePath)
	if err != nil {
		return nil, err
	}

	var tokens types.Token
	err = json.Unmarshal(file, &tokens)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func AccessToken() (string, error) {
	mu.Lock()

	tokens, err := loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %v", err)
	}

	if tokens.ExpiresIn > time.Now().Unix() {
		// Token is still valid
		mu.Unlock()
		return tokens.AccessToken, nil
	}

	mu.Unlock()
	return refreshAccessToken()
}

func saveTokens(tokens *types.Token) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(tokensFilePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func refreshAccessToken() (string, error) {
	mu.Lock()
	defer mu.Unlock()

	tokens, err := loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %v", err)
	}

	if tokens.ExpiresIn > time.Now().Unix() {
		// Token is still valid
		return tokens.AccessToken, nil
	}

	if tokens.RefreshToken == "" {
		return "", fmt.Errorf("refresh token not found")
	}

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", tokens.RefreshToken)

	req, err := http.NewRequest("POST", "https://api.canva.com/rest/v1/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	clientID := os.Getenv("CANVA_CLIENT_ID")
	clientSecret := os.Getenv("CANVA_CLIENT_SECRET")
	if clientID == "" {
		return "", fmt.Errorf("CANVA_CLIENT_ID not set")
	}
	if clientSecret == "" {
		return "", fmt.Errorf("CANVA_CLIENT_SECRET not set")
	}

	req.URL.RawQuery = form.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error refreshing access token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error refreshing access token: %s", body)
	}

	var newTokens types.Token
	err = json.NewDecoder(resp.Body).Decode(&newTokens)

	if err != nil {
		return "", err
	}

	// Update the expiration time and save the tokens
	newTokens.ExpiresIn = time.Now().Unix() + (newTokens.ExpiresIn - time.Now().Unix())
	err = saveTokens(&newTokens)
	if err != nil {
		return "", err
	}

	log.Println("Canva token refreshed successfully")
	return newTokens.AccessToken, nil
}

func StartCanvaTokenRefresher(interval time.Duration) {
	_, err := refreshAccessToken()
	if err != nil {
		log.Fatalf("Error refreshing token: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		_, err := refreshAccessToken()
		if err != nil {
			log.Printf("Error refreshing token: %v", err)
		} else {
			log.Println("Token refreshed successfully")
		}
	}
}
