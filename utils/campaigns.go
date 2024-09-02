package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/ethanhosier/mia-backend-go/prompts"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/sashabaranov/go-openai"
)

func Themes(themePrompt string, llmClient *LLMClient) ([]types.ThemeData, error) {
	completion, err := llmClient.OpenaiCompletion(themePrompt, openai.GPT4oMini)

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

func OptimalKeywords(keywords []types.GoogleAdsKeyword) (string, string) {

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

	primaryKeyword, secondaryKeyword := getOptimalKeyword(keywordSearchResults, filteredKeywords)

	return primaryKeyword, secondaryKeyword
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

func getOptimalKeyword(keywordSearchResults map[string]int, keywords []types.GoogleAdsKeyword) (primaryKeyword string, secondaryKeyword string) {
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

	primaryKeyword, secondaryKeyword = "", ""
	maxVal, secondMaxVal := float32(0), float32(0)

	for i, keyword := range keywords {
		score := 0.05*Normalize(float32(keyword.HighTopOfPageBid), allTopBids) +
			0.05*Normalize(float32(keyword.LowTopOfPageBid), allLowBids) +
			0.3*Normalize(float32(keyword.CompetitionIndex), allCompetitionIndexes) +
			0.5*Normalize(allInvertedKds[i], allInvertedKds)

		if score > maxVal {
			secondMaxVal = maxVal
			secondaryKeyword = primaryKeyword

			maxVal = score
			primaryKeyword = keyword.Keyword
		} else if score > secondMaxVal {
			secondMaxVal = score
			secondaryKeyword = keyword.Keyword
		}
	}

	return primaryKeyword, secondaryKeyword
}

func ResearchReportData(keyword string, llmClient *LLMClient) (types.ResearchReportData, error) {
	platforms := []string{"linkedIn", "facebook", "instagram", "google", "news"}

	ch := make(chan *types.PlatformResearchReport, len(platforms))
	wg := sync.WaitGroup{}
	wg.Add(len(platforms))

	for _, platform := range platforms {
		go func(platform string) {
			defer wg.Done()
			data, err := platformResearchReport(keyword, platform)
			if err != nil {
				log.Println("Error getting platform research report:", err)
				return
			}

			ch <- data
		}(platform)
	}
	wg.Wait()
	close(ch)

	platformResearchReports := []types.PlatformResearchReport{}
	for platformResearchReport := range ch {
		platformResearchReports = append(platformResearchReports, *platformResearchReport)
	}

	return types.ResearchReportData{
		PlatformResearchReports: platformResearchReports,
	}, nil
}

func platformResearchReport(keyword string, platform string) (*types.PlatformResearchReport, error) {
	resp, err := http.Get(SocialMediaFromKeywordScraperUrl + "?keyword=" + url.QueryEscape(keyword) + "&platform=" + platform + "&maxResults=5")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response types.SocialMediaFromKeywordResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &types.PlatformResearchReport{
		Platform: response.Platform,
		Posts:    response.Posts,
	}, nil
}

func CreateColorImage(hexColor string) ([]byte, error) {
	c, err := HexToColor(hexColor)
	if err != nil {
		return nil, err
	}

	// Create a new RGBA image with the desired size.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill the image with the color.
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)

	// Encode the image to a buffer.
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ColorsFromUrl(url string, llmClient *LLMClient) ([]string, error) {
	screenshotBase64, err := PageScreenshot(url)
	if err != nil {
		return nil, fmt.Errorf("error taking screenshot of page: %v", err)
	}

	resp, err := llmClient.OpenaiImageCompletion(prompts.ColorThemesPrompt, []string{screenshotBase64}, openai.GPT4o)
	if err != nil {
		return nil, err
	}

	var colors []string
	err = json.Unmarshal([]byte(resp), &colors)
	if err != nil {
		return nil, err
	}

	return colors, nil
}
