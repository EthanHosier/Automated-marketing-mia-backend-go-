package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/campaigns"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

type CampaignRequest struct {
	ID string `json:"id"`
}

func GenerateCampaigns(store storage.Storage, campaignClient *campaigns.CampaignClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}
		var req CampaignRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Now you can access the ID
		id := req.ID

		if id == "" {
			http.Error(w, "Campaign ID is required", http.StatusBadRequest)
			return
		}

		themes, err := campaignClient.GenerateThemesForUser(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		businessSummary, err := storage.Get[researcher.BusinessSummary](store, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		posts, researchReport, err := campaignClient.CampaignFrom(r.Context(), themes[0], businessSummary)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		postsResponses := []storage.Post{}
		for _, post := range posts {
			postsResponses = append(postsResponses, *post)
		}

		campaign := storage.Campaign{
			ID: id,
			Data: storage.CampaignData{
				ResearchReport: researchReport,
				Posts:          postsResponses,
				Theme:          themes[0].Theme,
				PrimaryKeyword: themes[0].PrimaryKeyword,
			},
		}

		err = storage.Store(store, campaign)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetCampaign(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Campaign ID is required", http.StatusBadRequest)
			return
		}

		campaign, err := storage.Get[storage.Campaign](store, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(campaign)
	}
}
