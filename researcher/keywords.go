package researcher

import (
	"fmt"
	"sync"
)

func (r *Researcher) GoogleAdsKeywordsData(keywords []string) ([]GoogleAdsKeyword, error) {
	googleAdsKeywordResponses, err := r.servicesClient.GoogleAdsKeywordsData(keywords)
	if err != nil {
		return nil, err
	}

	googleAdsKeywords := []GoogleAdsKeyword{}
	for _, keyword := range googleAdsKeywordResponses {
		googleAdsKeywords = append(googleAdsKeywords, GoogleAdsKeyword{
			Keyword:            keyword.Keyword,
			AvgMonthlySearches: keyword.AvgMonthlySearches,
			CompetitionLevel:   keyword.CompetitionLevel,
			CompetitionIndex:   keyword.CompetitionIndex,
			LowTopOfPageBid:    keyword.LowTopOfPageBid,
			HighTopOfPageBid:   keyword.HighTopOfPageBid,
		})
	}

	return googleAdsKeywords, nil
}

func (r *Researcher) OptimalKeywords(keywords []GoogleAdsKeyword) (string, string, error) {
	wg := sync.WaitGroup{}
	wg.Add(len(keywords))

	errorCh := make(chan error, len(keywords))
	ch := make(chan map[string]int, len(keywords))

	for _, keyword := range keywords {
		go func(keyword GoogleAdsKeyword) {
			defer wg.Done()
			volume, err := r.servicesClient.NumberOfSearchResultsFor(keyword.Keyword)

			if err == nil {
				ch <- map[string]int{keyword.Keyword: volume}
				return
			}

			errorCh <- fmt.Errorf("error getting search results for %v:%v", keyword.Keyword, err)
		}(keyword)
	}
	wg.Wait()
	close(ch)

	select {
	case err := <-errorCh:
		return "", "", err
	default:
	}

	keywordSearchResults := map[string]int{}
	for searchResult := range ch {
		for k, v := range searchResult {
			keywordSearchResults[k] = v
		}
	}

	filteredKeywords := []GoogleAdsKeyword{}
	for _, keyword := range keywords {
		if keywordSearchResults[keyword.Keyword] != -1 {
			filteredKeywords = append(filteredKeywords, keyword)
		}
	}

	primaryKeyword, secondaryKeyword := getOptimalKeyword(keywordSearchResults, filteredKeywords)

	return primaryKeyword, secondaryKeyword, nil
}

func getOptimalKeyword(keywordSearchResults map[string]int, keywords []GoogleAdsKeyword) (primaryKeyword string, secondaryKeyword string) {
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
		score := 0.05*normalize(float32(keyword.HighTopOfPageBid), allTopBids) +
			0.05*normalize(float32(keyword.LowTopOfPageBid), allLowBids) +
			0.3*normalize(float32(keyword.CompetitionIndex), allCompetitionIndexes) +
			0.5*normalize(allInvertedKds[i], allInvertedKds)

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

func normalize(x float32, nums []float32) float32 {
	min, max := minMax(nums)
	if min == max {
		return 0.0
	}
	return (x - min) / (max - min)
}

func minMax(nums []float32) (float32, float32) {
	min, max := nums[0], nums[0]
	for _, num := range nums {
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}
	return min, max
}

func invertedKd(volume int, results int) float32 {
	return float32(volume) / float32(results)
}
