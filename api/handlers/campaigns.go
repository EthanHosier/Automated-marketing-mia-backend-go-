package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/storage"
)

func GenerateCampaigns(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(nil)
	}
}
