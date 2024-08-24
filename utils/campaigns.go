package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

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

func OptimalKeyword(keywords []types.GoogleAdsKeyword) string {

	wg := sync.WaitGroup{}
	wg.Add(len(keywords))

	ch := make(chan map[string]int, len(keywords))

	for _, keyword := range keywords {
		go func(keyword types.GoogleAdsKeyword) {
			defer wg.Done()
			volume, err := searchResults(keyword.Keyword)

			if err == nil {
				ch <- map[string]int{keyword.Keyword: volume}
				return
			}

			log.Println("Error getting search results for", keyword.Keyword, ":", err)

		}(keyword)
	}
	wg.Wait()
	close(ch)

	keywordSearchResults := map[string]int{}
	for searchResult := range ch {
		for k, v := range searchResult {
			keywordSearchResults[k] = v
		}
	}

	filteredKeywords := []types.GoogleAdsKeyword{}
	for _, keyword := range keywords {
		if keywordSearchResults[keyword.Keyword] != -1 {
			filteredKeywords = append(filteredKeywords, keyword)
		}
	}

	optimalKeyword := getOptimalKeyword(keywordSearchResults, filteredKeywords)

	return optimalKeyword
}

func searchResults(keyword string) (int, error) {
	k := url.QueryEscape(keyword)
	resp, err := http.Get(SearchResultsUrl + k)

	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	var response types.SearchResultsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return -1, err
	}

	return response.SearchResults, nil
}

func invertedKd(volume int, results int) float32 {
	return float32(volume) / float32(results)
}

func getOptimalKeyword(keywordSearchResults map[string]int, keywords []types.GoogleAdsKeyword) string {
	allTopBids := []float32{}
	allLowBids := []float32{}
	allCompetitionIndexes := []float32{}
	allInvertedKds := []float32{}

	for _, keyword := range keywords {
		allTopBids = append(allTopBids, float32(keyword.HighTopOfPageBid))
		allLowBids = append(allLowBids, float32(keyword.LowTopOfPageBid))
		allCompetitionIndexes = append(allCompetitionIndexes, float32(keyword.CompetitionIndex))

		invertedKd := invertedKd(keyword.AvgMonthlySearches, keywordSearchResults[keyword.Keyword])
		allInvertedKds = append(allInvertedKds, invertedKd)
	}

	optimalKeyword, maxVal := "", float32(0)

	for i, keyword := range keywords {
		score := 0.05*Normalize(float32(keyword.HighTopOfPageBid), allTopBids) + 0.05*Normalize(float32(keyword.LowTopOfPageBid), allLowBids) + 0.3*Normalize(float32(keyword.CompetitionIndex), allCompetitionIndexes) + 0.5*Normalize(allInvertedKds[i], allInvertedKds)

		if score > maxVal {
			maxVal = score
			optimalKeyword = keyword.Keyword
		}
	}

	return optimalKeyword
}
