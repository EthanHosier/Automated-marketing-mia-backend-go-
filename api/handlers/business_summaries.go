package handlers

import (
	"encoding/json"
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

		// check user hasn't already generated business summaries or sitemap
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
		if !valid {
			http.Error(w, "User has already generated business summaries or sitemap", http.StatusBadRequest)
			return
		}

		sitemapWg := sync.WaitGroup{}
		sitemapWg.Add(1)

		go func() {
			defer sitemapWg.Done()

			urls, err := utils.Sitemap(req.Url, 15)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			uniqueUrls := utils.RemoveDuplicates(urls)
			embeddings, err := llmClient.OpenaiEmbeddings(uniqueUrls)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			log.Println("Storing sitemap for user", userID, "with", len(uniqueUrls), "unique URLs")
			err = store.StoreSitemap(userID, urls, embeddings)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}()

		summaries, err := utils.BusinessPageSummaries(req.Url, 15, llmClient)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(summaries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
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
