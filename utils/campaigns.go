package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ethanhosier/mia-backend-go/types"
)

func Themes(themePrompt string, llmClient *LLMClient) ([]types.ThemeData, error) {
	completion, err := llmClient.OpenaiCompletion(themePrompt)

	if err != nil {
		return nil, err
	}

	extractedArr := ExtractJsonObj(completion, SquareBracket)

	var themeData []types.ThemeData
	err = json.Unmarshal([]byte(extractedArr), &themeData)
	if err != nil {
		return nil, err
	}

	return themeData, nil
}

func GoogleAdsKeywordsData(keywords []string) ([]types.GoogleAdsKeyword, error) {
	queryKeywords := []string{}

	for _, keyword := range keywords {
		queryKeywords = append(queryKeywords, url.QueryEscape(keyword))
	}

	keywordsStr := strings.Join(queryKeywords, ",")

	resp, err := http.Get(GoogleAdsUrl + keywordsStr)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response types.GoogleAdsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Keywords, nil
}
