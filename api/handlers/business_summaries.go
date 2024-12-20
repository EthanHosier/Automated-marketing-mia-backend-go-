package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/ethanhosier/mia-backend-go/images"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
	"github.com/google/uuid"
)

func BusinessSummaries(store storage.Storage, rr researcher.Researcher, imageClient images.ImagesClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		var req BusinessSummariesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateBusinessSummariesRequest(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if alreadyHasSitemapOrBusinessSummary(store, userID) {
			http.Error(w, "User already has sitemap or business summary", http.StatusBadRequest)
			return
		}

		urls, businessSummaries, imageUrls, err := rr.BusinessSummary(req.Url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		notTooSmallUrls, err := imageClient.FilterTooSmallImages(imageUrls)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlIsValid := make([]bool, len(imageUrls))
		captionsList, err := imageClient.CaptionsForAll(notTooSmallUrls[:min(50, len(notTooSmallUrls))])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for i, captions := range captionsList {
			if len(captions) == 0 {
				urlIsValid[i] = false
			} else {
				urlIsValid[i] = true
			}
		}

		validUrls := []string{}
		for i, isValid := range urlIsValid {
			if isValid {
				validUrls = append(validUrls, notTooSmallUrls[i])
			}
		}

		validCaptions := [][]string{}
		for _, cl := range captionsList {
			if len(cl) > 0 {
				validCaptions = append(validCaptions, cl)
			}
		}

		// TODO: make this flat so all in one request
		embeddingsTaks := utils.DoAsyncList(validCaptions, func(captions []string) ([][]float32, error) {
			return rr.EmbeddingsFor(captions)
		})

		embeddings, err := utils.GetAsyncList(embeddingsTaks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		imgFeatures := []storage.ImageFeature{}
		for i, captions := range validCaptions {
			for ci, caption := range captions {
				imgFeatures = append(imgFeatures, storage.ImageFeature{
					ID:               uuid.New().String(),
					Feature:          caption,
					FeatureEmbedding: embeddings[i][ci],
					UserId:           userID,
					ImageUrl:         validUrls[i],
				})
			}
		}

		err = storage.StoreAll(store, imgFeatures...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u := []researcher.SitemapUrl{}
		for _, url := range urls {
			u = append(u, researcher.SitemapUrl{Url: url, ID: userID})
		}

		err = storage.StoreAll(store, u...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		businessSummaries.ID = userID
		err = storage.Store(store, *businessSummaries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(businessSummaries)
	}
}

func GetBusinessSummaries(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		businessSummary, err := storage.Get[researcher.BusinessSummary](store, userID)
		if err == storage.NotFoundError {
			http.Error(w, "Business summary not found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(businessSummary)
	}
}

func PatchBusinessSummaries(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		// TODO: validate the fields are of the correct type
		var updateFields map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updateFields); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := storage.Update[researcher.BusinessSummary](store, userID, updateFields)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode("Business summaries updated")
	}
}

func alreadyHasSitemapOrBusinessSummary(store storage.Storage, userID string) bool {
	checkWg := sync.WaitGroup{}
	checkWg.Add(1)

	hasSitemap := true

	go func() {
		defer checkWg.Done()
		businessSummary, err := storage.Get[researcher.BusinessSummary](store, userID)
		if err != nil || businessSummary.BusinessSummary == "" {
			hasSitemap = false
		}
	}()

	urls, err := storage.GetAll[researcher.SitemapUrl](store, map[string]string{"id": userID})
	if err != nil || len(urls) == 0 {
		hasSitemap = false
	}

	checkWg.Wait()

	return hasSitemap
}

// func saveSitemap(userID string, urls []string, llmClient *utils.LLMClient, store storage.Storage) error {
// 	if len(urls) == 0 {
// 		return errors.New("no URLs to save")
// 	}

// 	embeddings, err := llmClient.OpenaiEmbeddings(urls)
// 	if err != nil {
// 		return err
// 	}
// 	log.Println("Storing sitemap for user", userID, "with", len(urls), "unique URLs")
// 	err = store.StoreSitemap(userID, urls, embeddings)

// 	return err
// }
