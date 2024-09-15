package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/campaigns"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

func GenerateCampaigns(store storage.Storage, campaignClient *campaigns.CampaignClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		themes, err := campaignClient.GenerateThemesForUser(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		businessSummary, err := storage.Get[researcher.BusinessSummary](store, userID)
		templates, researchReport, err := campaignClient.CampaignFrom(themes[0], businessSummary)

		slog.Info("Generated campaign for user %s", userID, "with %d templates", len(templates))
		fmt.Printf("%+v", templates)

		type response struct {
			ResearchReport string `json:"research_report"`
		}

		resp := response{
			ResearchReport: *researchReport,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
