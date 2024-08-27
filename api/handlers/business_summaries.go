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

	valid := true

	go func() {
		defer checkWg.Done()
		_, err := store.GetBusinessSummary(userID)
		if err == nil {
			valid = false
		}
	}()

	_, err := store.GetSitemap(userID)
	if err == nil {
		valid = false
	}

	checkWg.Wait()

	return !valid
}

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
