package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/ethanhosier/mia-backend-go/utils"
)

const (
	maxBusinessSummaryUrls = 40
	maxFullPageScrapes     = 5
)

func BusinessSummaries2(store storage.Storage, llmClient *utils.LLMClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		var req types.BusinessSummariesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := types.ValidateBusinessSummariesRequest(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if alreadyHasSitemapOrBusinessSummary(store, userID) {
			http.Error(w, "User already has sitemap or business summary", http.StatusBadRequest)
			return
		}

		urls, err := utils.Sitemap(req.Url, 15)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		uniqueUrls := utils.RemoveDuplicates(urls)

		saveSitemapWg := sync.WaitGroup{}
		saveSitemapWg.Add(1)

		go func() {
			defer saveSitemapWg.Done()
			err := saveSitemap(userID, uniqueUrls, llmClient, store)

			if err != nil {
				fmt.Println("Error saving sitemap:", err)
			}
		}()

		sortedUrls, err := utils.SortURLsByProximity(uniqueUrls)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		imageUrls, bodyTexts, err := scrapeWebsitePages(sortedUrls[:min(maxBusinessSummaryUrls, len(sortedUrls))])

		for _, text := range bodyTexts {
			fmt.Println(text)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonTexts, err := json.Marshal(bodyTexts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		businessSummaries, err := utils.BusinessSummaryPoints(string(jsonTexts), llmClient)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		businessSummaryResponse := types.BusinessSummariesResponse{
			BusinessSummaries: types.BusinessSummary{
				BusinessName:    businessSummaries.BusinessName,
				BusinessSummary: businessSummaries.BusinessSummary,
				BrandVoice:      businessSummaries.BrandVoice,
				TargetRegion:    businessSummaries.TargetRegion,
				TargetAudience:  businessSummaries.TargetAudience,
			},
			ImageUrls: imageUrls,
		}

		store.StoreBusinessSummary(userID, *businessSummaries)

		saveSitemapWg.Wait()

		json.NewEncoder(w).Encode(businessSummaryResponse)
	}
}

func scrapeWebsitePages(urls []string) ([]string, []string, error) {
	n := len(urls)

	pageWg := sync.WaitGroup{}
	pageWg.Add(n)

	pageCh := make(chan types.BodyContentsScrapeResponse, n)

	for _, url := range urls {
		go func(url string) {
			defer pageWg.Done()

			pageContents, err := utils.PageContentsScrape(url)
			if err != nil {
				fmt.Println("Error scraping page contents:", err)
				return
			}
			pageCh <- *pageContents
		}(url)
	}
	pageWg.Wait()
	close(pageCh)

	imageSet := make(map[string]struct{})
	pageContents := []string{}

	for contents := range pageCh {
		for _, imageUrl := range contents.ImageUrls {
			imageSet[imageUrl] = struct{}{}
		}
		pageContents = append(pageContents, utils.StringifyWebsiteData(contents.Contents))
	}

	images := make([]string, 0, len(imageSet))
	for imageUrl := range imageSet {
		images = append(images, imageUrl)
	}

	return images, pageContents, nil
}

func BusinessSummaries(store storage.Storage, llmClient *utils.LLMClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		var req types.BusinessSummariesRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := types.ValidateBusinessSummariesRequest(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if alreadyHasSitemapOrBusinessSummary(store, userID) {
			http.Error(w, "User already has sitemap or business summary", http.StatusBadRequest)
			return
		}

		sitemapWg := sync.WaitGroup{}
		sitemapWg.Add(1)

		go func() {
			defer sitemapWg.Done()
			err := scrapeSitemap(req, userID, llmClient, store)

			if err != nil {
				fmt.Println("Error scraping sitemap:", err)
			}
		}()

		businessSummaries, err := scrapeBusinessSummaries(req.Url, llmClient)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Storing business summaries for user", userID)
		err = store.StoreBusinessSummary(userID, *businessSummaries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := types.BusinessSummariesResponse{
			BusinessSummaries: *businessSummaries,
		}

		sitemapWg.Wait()
		json.NewEncoder(w).Encode(resp)
	}
}

func GetBusinessSummaries(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		businessSummary, err := store.GetBusinessSummary(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := types.BusinessSummariesResponse{
			BusinessSummaries: types.BusinessSummary{
				BusinessName:    "",
				BusinessSummary: businessSummary.BusinessSummary,
				BrandVoice:      businessSummary.BrandVoice,
				TargetRegion:    businessSummary.TargetRegion,
				TargetAudience:  businessSummary.TargetAudience,
			},
		}

		json.NewEncoder(w).Encode(resp)
	}
}

func alreadyHasSitemapOrBusinessSummary(store storage.Storage, userID string) bool {
	checkWg := sync.WaitGroup{}
	checkWg.Add(1)

	hasSitemap := true

	go func() {
		defer checkWg.Done()
		businessSummary, err := store.GetBusinessSummary(userID)
		if err != nil || businessSummary.BusinessSummary == "" {
			hasSitemap = false
		}
	}()

	urls, err := store.GetSitemap(userID)
	if len(urls) == 0 || err != nil {
		hasSitemap = false
	}

	checkWg.Wait()

	return hasSitemap
}

// Deprecated
func scrapeSitemap(req types.BusinessSummariesRequest, userID string, llmClient *utils.LLMClient, store storage.Storage) error {
	urls, err := utils.Sitemap(req.Url, 15)
	if err != nil {
		return err
	}

	uniqueUrls := utils.RemoveDuplicates(urls)
	embeddings, err := llmClient.OpenaiEmbeddings(uniqueUrls)

	if err != nil {
		return err
	}

	log.Println("Storing sitemap for user", userID, "with", len(uniqueUrls), "unique URLs")
	err = store.StoreSitemap(userID, urls, embeddings)

	return err
}

func saveSitemap(userID string, urls []string, llmClient *utils.LLMClient, store storage.Storage) error {
	embeddings, err := llmClient.OpenaiEmbeddings(urls)
	if err != nil {
		return err
	}
	log.Println("Storing sitemap for user", userID, "with", len(urls), "unique URLs")
	err = store.StoreSitemap(userID, urls, embeddings)

	return err
}

func scrapeBusinessSummaries(url string, llmClient *utils.LLMClient) (*types.BusinessSummary, error) {
	summaries, err := utils.BusinessPageSummaries(url, 15, llmClient)

	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(summaries)
	if err != nil {
		return nil, err
	}

	jsonString := string(jsonData)

	businessSummaries, err := utils.BusinessSummaryPoints(jsonString, llmClient)

	if err != nil {
		log.Println("Error getting business summary points:", err, ". Trying again (1st retry)")
		businessSummaries, err = utils.BusinessSummaryPoints(jsonString, llmClient)

		if err != nil {
			log.Println("Error getting business summary points:", err, ". Trying again (2nd retry)")
			businessSummaries, err = utils.BusinessSummaryPoints(jsonString, llmClient)

			if err != nil {
				return nil, err
			}
		}
	}

	return businessSummaries, nil
}
