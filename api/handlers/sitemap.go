package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/ethanhosier/mia-backend-go/utils"
)

func GetSitemap(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		sitemap, err := store.GetSitemap(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urls := []string{}
		for _, url := range sitemap {
			urls = append(urls, url.Url)
		}

		resp := types.SitemapResponse{
			Urls: urls,
		}

		json.NewEncoder(w).Encode(resp)
	}
}
