package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/prompts"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/ethanhosier/mia-backend-go/utils"
	"github.com/sashabaranov/go-openai"
)

func GenerateCampaigns(store storage.Storage, llmClient *utils.LLMClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		businessSummary, err := store.GetBusinessSummary(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		themes, err := generateThemes(businessSummary, req.TargetAudienceLocation, req.Instructions, req.Backlink, []string{}, llmClient)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		campaign, err := campaignFromTheme(themes[0], businessSummary, llmClient, store)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("generated campaign", campaign)
		resp := types.GenerateCampaignsResponse{}

		json.NewEncoder(w).Encode(resp)
	}
}

func generateThemes(businessSummary types.StoredBusinessSummary, targetAudienceLocation string, additionalInstructions string, backlink string, imageDescriptions []string, llmClient *utils.LLMClient) ([]types.ThemeData, error) {
	themePrompt := prompts.ThemePrompt(businessSummary, targetAudienceLocation, additionalInstructions, backlink, imageDescriptions)
	themes, err := utils.Themes(themePrompt, llmClient)

	if err != nil {
		log.Println("Error generating themes:", err, ". Trying again (1st retry)")
		themes, err = utils.Themes(themePrompt, llmClient)

		if err != nil {
			log.Println("Error generating themes:", err, ". Trying again (2nd retry)")
			themes, err = utils.Themes(themePrompt, llmClient)

			if err != nil {
				return nil, errors.New("error generating themes")
			}
		}
	}

	return themes, nil
}

func campaignFromTheme(theme types.ThemeData, businessSummary types.StoredBusinessSummary, llmClient *utils.LLMClient, store storage.Storage) (string, error) {
	log.Printf("Generating campaign from theme \"%v\"\n", theme.Theme)

	url, err := campaignUrl(theme.UrlDescription, store, llmClient)
	if err != nil {
		return "", fmt.Errorf("error getting URL for campaign: %w", err)
	}

	primaryKeyword, secondaryKeyword, err := campaignChosenKewords(theme.Keywords)
	if err != nil {
		return "", fmt.Errorf("error getting keywords for campaign: %w", err)
	}

	template, err := chosenTemplate(theme.FacebookPostDescription, llmClient, store)
	if err != nil {
		return "", fmt.Errorf("error getting template for campaign: %w", err)
	}

	researchReportData, err := utils.ResearchReportData(primaryKeyword, llmClient)
	if err != nil {
		return "", fmt.Errorf("error getting research report data: %w", err)
	}

	for _, platformResearchReport := range researchReportData.PlatformResearchReports {
		fmt.Println(platformResearchReport.Platform)
		for _, post := range platformResearchReport.SummarisedPosts {
			fmt.Printf("%+v\n\n", post)
		}
	}

	researchReportPrompt := prompts.ResearchReportPrompt(primaryKeyword, researchReportData)
	researchReport, err := llmClient.OpenaiCompletion(researchReportPrompt, openai.GPT4o)
	if err != nil {
		return "", fmt.Errorf("error generating research report: %w", err)
	}

	fmt.Printf("URL: %v\nPrimary Keyword: %v\nSecondary Keyword: %v\nTemplate: %+v\n\n\n\n Reserch report: %v\n", url, primaryKeyword, secondaryKeyword, *template, researchReport)

	scrapedPageBody, err := utils.PageTextContents(url)
	if err != nil {
		return "", fmt.Errorf("error getting page text contents: %w", err)
	}

	summarisedPageBody, err := llmClient.OpenaiCompletion(prompts.BacklinkUrlPageSummary+scrapedPageBody, openai.GPT4oMini)
	if err != nil {
		return "", fmt.Errorf("error summarising page body: %w", err)
	}

	templatePrompt := prompts.TemplatePrompt("linkedIn", businessSummary, theme.Theme, primaryKeyword, secondaryKeyword, url, summarisedPageBody, researchReportData, template.Fields)

	fmt.Println("template prompt: ", templatePrompt)

	templateCompletion, err := llmClient.OpenaiCompletion(templatePrompt, openai.GPT4o)
	if err != nil {
		return "", fmt.Errorf("error generating template completion: %w", err)
	}

	extractedTemplate := utils.ExtractJsonObj(templateCompletion, utils.CurlyBracket)

	var populatedTemplate types.PopulatedTemplate
	err = json.Unmarshal([]byte(extractedTemplate), &populatedTemplate)

	if err != nil {
		return "", fmt.Errorf("error unmarshalling populated template: %w", err)
	}

	err = utils.PopulateTemplate(*template, populatedTemplate)
	if err != nil {
		return "", fmt.Errorf("error populating template: %w", err)
	}

	return "", nil
}

func campaignUrl(urlDescription string, store storage.Storage, llmClient *utils.LLMClient) (string, error) {
	urlEmbedding, err := llmClient.OpenaiEmbedding(urlDescription)
	if err != nil {
		return "", fmt.Errorf("error getting embedding for URL description: %w", err)
	}

	nearestUrl, err := store.GetNearestUrl(urlEmbedding)
	if err != nil {
		return "", fmt.Errorf("error getting nearest URL: %w", err)
	}

	return nearestUrl, nil
}

func campaignChosenKewords(keywords []string) (string, string, error) {
	adsKeywordData, err := utils.GoogleAdsKeywordsData(keywords)

	if err != nil {
		return "", "", fmt.Errorf("error getting Google Ads data: %w", err)
	}

	primaryKeyword, secondaryKeyword := utils.OptimalKeywords(adsKeywordData)

	return primaryKeyword, secondaryKeyword, nil
}

func chosenTemplate(templateDescription string, llmClient *utils.LLMClient, store storage.Storage) (*types.NearestTemplateResponse, error) {
	embedding, err := llmClient.OpenaiEmbedding(templateDescription)
	if err != nil {
		return nil, fmt.Errorf("error getting embedding for optimal keyword: %w", err)
	}

	template, err := store.GetNearestTemplate(embedding)
	if err != nil {
		return nil, fmt.Errorf("error getting nearest template: %w", err)
	}

	return template, nil
}
