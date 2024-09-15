package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

func GetSitemap(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		// TODO: define sitemap type
		sitemap, err := storage.GetAll[researcher.SitemapUrl](store, map[string]string{"id": userID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urls := []researcher.SitemapUrl{}
		for _, url := range sitemap {
			urls = append(urls, url)
		}

		type sitemapResponse struct {
			Urls []researcher.SitemapUrl `json:"urls"`
		}

		json.NewEncoder(w).Encode(sitemapResponse{Urls: urls})
	}
}
