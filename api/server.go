package api

import (
	"encoding/json"
	"errors"
	"fmt"
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

	themes, err := s.generateThemes(businessSummary, req.TargetAudienceLocation, req.Instructions, req.Backlink, []string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	campaign, err := s.campaignFromTheme(themes[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("generated campaign", campaign)
	resp := types.GenerateCampaignsResponse{}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) generateThemes(businessSummary types.StoredBusinessSummary, targetAudienceLocation string, additionalInstructions string, backlink string, imageDescriptions []string) ([]types.ThemeData, error) {
	themePrompt := prompts.ThemePrompt(businessSummary, targetAudienceLocation, additionalInstructions, backlink, imageDescriptions)
	themes, err := utils.Themes(themePrompt, s.llmClient)

	if err != nil {
		log.Println("Error generating themes:", err, ". Trying again (1st retry)")
		themes, err = utils.Themes(themePrompt, s.llmClient)

		if err != nil {
			log.Println("Error generating themes:", err, ". Trying again (2nd retry)")
			themes, err = utils.Themes(themePrompt, s.llmClient)

			if err != nil {
				return nil, errors.New("error generating themes")
			}
		}
	}

	return themes, nil
}

func (s *Server) campaignFromTheme(theme types.ThemeData) (string, error) {
	log.Printf("Generating campaign from theme \"%v\"\n", theme.Theme)

	url, err := s.campaignUrl(theme.UrlDescription)
	if err != nil {
		return "", fmt.Errorf("error getting URL for campaign: %w", err)
	}

	primaryKeyword, secondaryKeyword, err := s.campaignChosenKewords(theme.Keywords)
	if err != nil {
		return "", fmt.Errorf("error getting keywords for campaign: %w", err)
	}

	template, err := s.chosenTemplate(theme.FacebookPostDescription)
	if err != nil {
		return "", fmt.Errorf("error getting template for campaign: %w", err)
	}

	researchReportData, err := utils.ResearchReportData(primaryKeyword, s.llmClient)
	if err != nil {
		return "", fmt.Errorf("error getting research report data: %w", err)
	}

	researchReportPrompt := prompts.ResearchReportPrompt(primaryKeyword, researchReportData)
	researchReport, err := s.llmClient.OpenaiCompletion(researchReportPrompt)
	if err != nil {
		return "", fmt.Errorf("error generating research report: %w", err)
	}

	fmt.Printf("URL: %v\nPrimary Keyword: %v\nSecondary Keyword: %v\nTemplate: %+v\n\n\n\n Reserch report: %v\n", url, primaryKeyword, secondaryKeyword, template, researchReport)

	return "", nil
}

func (s *Server) campaignUrl(urlDescription string) (string, error) {
	urlEmbedding, err := s.llmClient.OpenaiEmbedding(urlDescription)
	if err != nil {
		return "", fmt.Errorf("error getting embedding for URL description: %w", err)
	}

	nearestUrl, err := s.store.GetNearestUrl(urlEmbedding)
	if err != nil {
		return "", fmt.Errorf("error getting nearest URL: %w", err)
	}

	return nearestUrl, nil
}

func (s *Server) campaignChosenKewords(keywords []string) (string, string, error) {
	adsKeywordData, err := utils.GoogleAdsKeywordsData(keywords)

	if err != nil {
		return "", "", fmt.Errorf("error getting Google Ads data: %w", err)
	}

	primaryKeyword, secondaryKeyword := utils.OptimalKeywords(adsKeywordData)

	return primaryKeyword, secondaryKeyword, nil
}

func (s *Server) chosenTemplate(templateDescription string) (*types.NearestTemplateResponse, error) {
	embedding, err := s.llmClient.OpenaiEmbedding(templateDescription)
	if err != nil {
		return nil, fmt.Errorf("error getting embedding for optimal keyword: %w", err)
	}

	template, err := s.store.GetNearestTemplate(embedding)
	if err != nil {
		return nil, fmt.Errorf("error getting nearest template: %w", err)
	}

	return template, nil
}
