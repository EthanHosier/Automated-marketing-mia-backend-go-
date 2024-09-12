package researcher

import (
	"encoding/json"
	"testing"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/services"
	"github.com/stretchr/testify/assert"
)

func TestGoogleAdsKeywordsData(t *testing.T) {
	// given
	var (
		httpClient = http.MockHttpClient{}

		servicesClient   = services.NewServicesClient(&httpClient)
		researcherClient = New(servicesClient, nil)

		keywords     = []string{"keyword1", "keyword2"}
		keywordsResp = []services.GoogleAdsKeywordResponse{
			{Keyword: "keyword1", AvgMonthlySearches: 1000, CompetitionLevel: "High", CompetitionIndex: 1, LowTopOfPageBid: 10, HighTopOfPageBid: 20},
			{Keyword: "keyword2", AvgMonthlySearches: 500, CompetitionLevel: "Medium", CompetitionIndex: 2, LowTopOfPageBid: 5, HighTopOfPageBid: 15},
		}

		keywordsJson, _ = json.Marshal(keywordsResp)
	)

	httpClient.WillReturnBody("GET", services.GoogleAdsUrl+"keyword1,keyword2", `{"keywords":`+string(keywordsJson)+`}`)

	// when
	result, err := researcherClient.GoogleAdsKeywordsData(keywords)

	// then
	assert.NoError(t, err)
	assert.Equal(t, []GoogleAdsKeyword{
		{Keyword: "keyword1", AvgMonthlySearches: 1000, CompetitionLevel: "High", CompetitionIndex: 1, LowTopOfPageBid: 10, HighTopOfPageBid: 20},
		{Keyword: "keyword2", AvgMonthlySearches: 500, CompetitionLevel: "Medium", CompetitionIndex: 2, LowTopOfPageBid: 5, HighTopOfPageBid: 15},
	}, result)
}

func TestOptimalKeywords(t *testing.T) {
	// given
	var (
		httpClient = http.MockHttpClient{}

		servicesClient   = services.NewServicesClient(&httpClient)
		researcherClient = New(servicesClient, nil)

		keywords = []GoogleAdsKeyword{
			{Keyword: "keyword1", AvgMonthlySearches: 1000, CompetitionLevel: "High", CompetitionIndex: 1, LowTopOfPageBid: 10, HighTopOfPageBid: 20},
			{Keyword: "keyword2", AvgMonthlySearches: 500, CompetitionLevel: "Medium", CompetitionIndex: 2, LowTopOfPageBid: 5, HighTopOfPageBid: 15},
			{Keyword: "keyword3", AvgMonthlySearches: 2000, CompetitionLevel: "Low", CompetitionIndex: 3, LowTopOfPageBid: 2, HighTopOfPageBid: 10},
		}
	)

	httpClient.WillReturnBody("GET", services.SearchResultsUrl+"keyword1", `{"searchResults": 1000000}`)
	httpClient.WillReturnBody("GET", services.SearchResultsUrl+"keyword2", `{"searchResults": 5000000}`)
	httpClient.WillReturnBody("GET", services.SearchResultsUrl+"keyword3", `{"searchResults": 10000000}`)

	// when
	primaryKeyword, secondaryKeyword, err := researcherClient.OptimalKeywords(keywords)

	// then
	assert.NoError(t, err)
	assert.Equal(t, "keyword1", primaryKeyword)
	assert.Equal(t, "keyword3", secondaryKeyword)
}

func TestGetOptimalKeyword(t *testing.T) {
	var (
		keywordSearchResults = map[string]int{
			"keyword1": 1000,
			"keyword2": 500,
			"keyword3": 2000,
		}

		keywords = []GoogleAdsKeyword{
			{Keyword: "keyword1", AvgMonthlySearches: 1000, CompetitionLevel: "High", CompetitionIndex: 1, LowTopOfPageBid: 10, HighTopOfPageBid: 20},
			{Keyword: "keyword2", AvgMonthlySearches: 500, CompetitionLevel: "Medium", CompetitionIndex: 2, LowTopOfPageBid: 5, HighTopOfPageBid: 15},
			{Keyword: "keyword3", AvgMonthlySearches: 2000, CompetitionLevel: "Low", CompetitionIndex: 3, LowTopOfPageBid: 7, HighTopOfPageBid: 25},
		}

		tests = []struct {
			name                 string
			keywordSearchResults map[string]int
			keywords             []GoogleAdsKeyword
			expectedPrimary      string
			expectedSecondary    string
		}{
			{
				name:                 "Basic Test",
				keywordSearchResults: keywordSearchResults,
				keywords:             keywords,
				expectedPrimary:      "keyword3",
				expectedSecondary:    "keyword2",
			},
			{
				name:                 "Empty Keyword List",
				keywordSearchResults: keywordSearchResults,
				keywords:             []GoogleAdsKeyword{},
				expectedPrimary:      "",
				expectedSecondary:    "",
			},
			{
				name:                 "Empty Search Results",
				keywordSearchResults: map[string]int{},
				keywords:             keywords,
				expectedPrimary:      "keyword3",
				expectedSecondary:    "keyword2",
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary, secondary := getOptimalKeyword(tt.keywordSearchResults, tt.keywords)
			if primary != tt.expectedPrimary {
				t.Errorf("expected primary %v, got %v", tt.expectedPrimary, primary)
			}
			if secondary != tt.expectedSecondary {
				t.Errorf("expected secondary %v, got %v", tt.expectedSecondary, secondary)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		x        float32
		nums     []float32
		expected float32
	}{
		{x: 2, nums: []float32{1, 2, 3}, expected: 0.5},
		{x: 1, nums: []float32{1, 1, 1}, expected: 0},
		{x: 3, nums: []float32{1, 2, 3}, expected: 1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := normalize(tt.x, tt.nums)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInvertedKd(t *testing.T) {
	tests := []struct {
		volume   int
		results  int
		expected float32
	}{
		{volume: 1000, results: 1000, expected: 1},
		{volume: 500, results: 1000, expected: 0.5},
		{volume: 2000, results: 1000, expected: 2},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := invertedKd(tt.volume, tt.results)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
