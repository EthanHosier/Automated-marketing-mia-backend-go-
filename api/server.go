package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/ethanhosier/mia-backend-go/prompts"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/ethanhosier/mia-backend-go/utils"
)

type Server struct {
	listenAddr string
	store      storage.Storage
	router     *http.ServeMux
	llmClient  *utils.LLMClient
}

func NewServer(listenAddr string, store storage.Storage, llmClient *utils.LLMClient) *Server {
	s := &Server{listenAddr: listenAddr, store: store, router: http.NewServeMux(), llmClient: llmClient}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("POST /business-summaries", s.businessSummaries)
	s.router.HandleFunc("GET /business-summaries", s.getBusinessSummaries)

	s.router.HandleFunc("GET /sitemap", s.getSitemap)

	s.router.HandleFunc("POST /campaigns", s.generateCampaigns)

}

func (s *Server) Start() error {
	stack := CreateMiddlewareStack(
		Auth,
		Logging,
	)

	return http.ListenAndServe(s.listenAddr, stack(s.router))
}

func (s *Server) businessSummaries(w http.ResponseWriter, r *http.Request) {
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
		_, err := s.store.GetBusinessSummary(userID)
		if err == nil {
			valid = false
		}
	}()

	_, err := s.store.GetSitemap(userID)
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
		embeddings, err := s.llmClient.OpenaiEmbeddings(uniqueUrls)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Storing sitemap for user", userID, "with", len(uniqueUrls), "unique URLs")
		err = s.store.StoreSitemap(userID, urls, embeddings)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}()

	summaries, err := utils.BusinessPageSummaries(req.Url, 15, s.llmClient)

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

	businessSummaries, err := utils.BusinessSummaryPoints(jsonString, s.llmClient)

	if err != nil {
		log.Println("Error getting business summary points:", err, ". Trying again (1st retry)")
		businessSummaries, err = utils.BusinessSummaryPoints(jsonString, s.llmClient)

		if err != nil {
			log.Println("Error getting business summary points:", err, ". Trying again (2nd retry)")
			businessSummaries, err = utils.BusinessSummaryPoints(jsonString, s.llmClient)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	log.Println("Storing business summaries for user", userID)
	err = s.store.StoreBusinessSummary(userID, *businessSummaries)
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

func (s *Server) getBusinessSummaries(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(utils.UserIdKey).(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	businessSummary, err := s.store.GetBusinessSummary(userID)
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

func (s *Server) getSitemap(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(utils.UserIdKey).(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	sitemap, err := s.store.GetSitemap(userID)
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

func (s *Server) generateCampaigns(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(utils.UserIdKey).(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	var req types.GenerateCampaignsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	businessSummary, err := s.store.GetBusinessSummary(userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	themePrompt := prompts.ThemePrompt(businessSummary, req.TargetAudienceLocation, []string{"https://example.com"}, req.Instructions, req.Backlink, []string{})
	themes, err := utils.Themes(themePrompt, s.llmClient)

	if err != nil {
		log.Println("Error generating themes:", err, ". Trying again (1st retry)")
		themes, err = utils.Themes(themePrompt, s.llmClient)

		if err != nil {
			log.Println("Error generating themes:", err, ". Trying again (2nd retry)")
			themes, err = utils.Themes(themePrompt, s.llmClient)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	resp := types.GenerateCampaignsResponse{
		Themes: themes,
	}

	json.NewEncoder(w).Encode(resp)
}
