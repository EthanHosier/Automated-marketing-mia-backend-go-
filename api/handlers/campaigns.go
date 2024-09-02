package handlers

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
	"github.com/sashabaranov/go-openai"
)

func GenerateCampaigns(store storage.Storage, llmClient *utils.LLMClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(utils.UserIdKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		pageContents := []types.BodyContentsScrapeResponse{}
		pageContentsWg := sync.WaitGroup{}
		pageContentsWg.Add(1)

		go func() {
			defer pageContentsWg.Done()
			var err error
			pageContents, err = candidatePages(userID, store)
			if err != nil {
				log.Println("Error getting candidate pages:", err)
			}
		}()

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

		pageContentsWg.Wait()

		themes, err := generateThemes(pageContents, businessSummary, req.TargetAudienceLocation, req.Instructions, req.Backlink, []string{}, llmClient)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		campaign, err := campaignFromTheme(themes[0], businessSummary, llmClient, store)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("\n\n\n\n")
		log.Println(len(campaign.Posts))
		log.Println("generated campaign %+v", campaign)
		resp := types.GenerateCampaignsResponse{}

		json.NewEncoder(w).Encode(resp)
	}
}

func generateThemes(pageContents []types.BodyContentsScrapeResponse, businessSummary types.StoredBusinessSummary, targetAudienceLocation string, additionalInstructions string, backlink string, imageDescriptions []string, llmClient *utils.LLMClient) ([]types.ThemeData, error) {

	themePrompt := prompts.ThemePrompt(pageContents, businessSummary, targetAudienceLocation, additionalInstructions, backlink, imageDescriptions)
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

func campaignFromTheme(theme types.ThemeData, businessSummary types.StoredBusinessSummary, llmClient *utils.LLMClient, store storage.Storage) (*types.GeneratedCampaign, error) {
	log.Printf("Generating campaign from theme \"%v\"\n", theme.Theme)
	url := theme.Url

	// Scrape the text of the given web page
	scrapedPageBodyCh := make(chan string, 1)
	go func() {
		scrapedPageText, err := utils.PageTextContents(url)
		if err != nil {
			fmt.Errorf("error getting page text contents: %w", err)
			scrapedPageBodyCh <- ""
			return
		}
		scrapedPageBodyCh <- scrapedPageText
	}()

	// Extract the main features of the page (including images)
	scrapedPageContentsCh := make(chan types.BodyContentsScrapeResponse, 1)
	go func() {
		pageContents, err := utils.PageContentsScrape(url)
		if err != nil {
			fmt.Errorf("error scraping page contents: %w", err)
			scrapedPageContentsCh <- types.BodyContentsScrapeResponse{}
			return
		}
		scrapedPageContentsCh <- *pageContents
	}()

	// Choose the template for the campaign
	templateCh := make(chan []types.NearestTemplateResponse, 1)
	go func() {
		templates, err := chosenTemplates(5, store)
		if err != nil {
			fmt.Errorf("error getting template for campaign: %w", err)
			templateCh <- nil
			return
		}
		templateCh <- templates
	}()

	primaryKeyword, secondaryKeyword, err := campaignChosenKewords(theme.Keywords)
	if err != nil {
		return nil, fmt.Errorf("error getting keywords for campaign: %w", err)
	}

	researchReportData, err := utils.ResearchReportData(primaryKeyword, llmClient)
	if err != nil {
		return nil, fmt.Errorf("error getting research report data: %w", err)
	}

	// Generate Research Report
	researchReportCh := make(chan string, 1)
	go func() {
		researchReportPrompt := prompts.ResearchReportPrompt(primaryKeyword, researchReportData)
		researchReport, err := llmClient.OpenaiCompletion(researchReportPrompt, openai.GPT4oMini)
		if err != nil {
			log.Println("error generating research report: %w", err)
			researchReportCh <- ""
			return
		}
		researchReportCh <- researchReport
	}()

	// Gather the template to be used for the campaign
	templates := <-templateCh
	scrapedPageBody := <-scrapedPageBodyCh
	scrapedPageContents := <-scrapedPageContentsCh

	fmt.Printf("Scraped page body contents: %+v\n", scrapedPageContents)

	platforms := []string{"instagram", "facebook", "twitterX", "linkedin", "whatsapp"}

	templateResultCh := make(chan *types.TemplateAndCaption, len(platforms))
	templateResultWg := sync.WaitGroup{}
	templateResultWg.Add(len(platforms))

	for i, platform := range platforms {
		go func(platform string) {
			resp, err := processTemplate(platform, businessSummary, theme, primaryKeyword, secondaryKeyword, url, scrapedPageBody, researchReportData, templates[i], scrapedPageContents, llmClient)
			if err != nil {
				log.Printf("Error processing template %v\n\n", err)
				templateResultCh <- nil
			}
			templateResultCh <- resp
		}(platform)
	}

	templateResultWg.Wait()
	close(templateResultCh)

	templateAndCaptions := []types.TemplateAndCaption{}
	for templateAndCaption := range templateResultCh {
		templateAndCaptions = append(templateAndCaptions, *templateAndCaption)
	}

	return &types.GeneratedCampaign{
		Posts:            templateAndCaptions,
		PrimaryKeyword:   primaryKeyword,
		SecondaryKeyword: secondaryKeyword,
		Theme:            theme.Theme,
	}, nil
}

func processTemplate(platform string, businessSummary types.StoredBusinessSummary, theme types.ThemeData, primaryKeyword string, secondaryKeyword string, url string, scrapedPageBody string, researchReportData types.ResearchReportData, template types.NearestTemplateResponse, pageContents types.BodyContentsScrapeResponse, llmClient *utils.LLMClient) (*types.TemplateAndCaption, error) {
	templatePrompt := prompts.TemplatePrompt("instagram", businessSummary, theme.Theme, primaryKeyword, secondaryKeyword, url, scrapedPageBody,
		researchReportData, template.Fields, template.ColorFields)

	templateCompletion, err := llmClient.OpenaiCompletion(templatePrompt, openai.GPT4oMini)
	if err != nil {
		return nil, fmt.Errorf("error generating template completion: %w", err)
	}
	extractedTemplate := utils.ExtractJsonObj(templateCompletion, utils.CurlyBracket)

	var populatedTemplate types.PopulatedTemplate
	err = json.Unmarshal([]byte(extractedTemplate), &populatedTemplate)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling populated template: %w.\n Template completion was %+v\n Extracted template was %+v", err, templateCompletion, extractedTemplate)
	}

	imageFields := []types.PopulatedField{}
	textFields := []types.PopulatedField{}

	for i := range populatedTemplate.Fields {
		if populatedTemplate.Fields[i].Type == types.ImageType {
			imageFields = append(imageFields, populatedTemplate.Fields[i])
			continue
		}

		if populatedTemplate.Fields[i].Type == types.TextType {
			textFields = append(textFields, populatedTemplate.Fields[i])
			continue
		}
	}

	campaignDetailsStr := fmt.Sprintf("Primary keyword: %v\nSecondary keyword: %v\nURL: %v\nTheme: %v\nTemplate Description: %v", primaryKeyword, secondaryKeyword, url, theme.Theme, theme.ImageCanvaTemplateDescription)

	bestImages, err := utils.PickBestImages(pageContents.ImageUrls, campaignDetailsStr, imageFields, llmClient)
	if err != nil {
		return nil, fmt.Errorf("error picking best images: %w", err)
	}

	imageFields, err = utils.UploadImageAssets(imageFields, bestImages)
	if err != nil {
		return nil, fmt.Errorf("error uploading image assets: %w", err)
	}

	colorFields, err := utils.UploadColorAssets(populatedTemplate.ColorFields)
	if err != nil {
		return nil, fmt.Errorf("error uploading color assets: %w", err)
	}

	log.Printf("Populated template: %+v\n", populatedTemplate)

	result, err := utils.PopulateTemplate(template.ID, imageFields, textFields, colorFields)
	if err != nil {
		return nil, fmt.Errorf("error populating template: %w", err)
	}

	return &types.TemplateAndCaption{
		TemplateResult: *result,
		Caption:        populatedTemplate.Caption,
	}, nil
}

func campaignChosenKewords(keywords []string) (string, string, error) {
	adsKeywordData, err := utils.GoogleAdsKeywordsData(keywords)

	if err != nil {
		return "", "", fmt.Errorf("error getting Google Ads data: %w", err)
	}

	primaryKeyword, secondaryKeyword := utils.OptimalKeywords(adsKeywordData)

	return primaryKeyword, secondaryKeyword, nil
}

func chosenTemplates(numTemplates int, store storage.Storage) ([]types.NearestTemplateResponse, error) {
	templates, err := store.GetRandomTemplates(numTemplates)
	if err != nil {
		return nil, fmt.Errorf("error getting nearest template: %w", err)
	}

	return templates, nil
}

func candidatePages(userID string, store storage.Storage) ([]types.BodyContentsScrapeResponse, error) {
	randomUrls, err := store.GetRandomUrls(userID, 5)
	if err != nil {
		return nil, fmt.Errorf("error getting random URLs: %w", err)
	}

	pageContentsWg := sync.WaitGroup{}
	pageContentsWg.Add(len(randomUrls))

	pageContentsCh := make(chan types.BodyContentsScrapeResponse, len(randomUrls))

	for _, url := range randomUrls {
		go func(url string) {
			defer pageContentsWg.Done()

			pageContents, err := utils.PageContentsScrape(url)
			if err != nil {
				log.Println("Error scraping page contents:", err)
				return
			}

			pageContentsCh <- *pageContents
		}(url)
	}

	pageContentsWg.Wait()
	close(pageContentsCh)

	pageContents := []types.BodyContentsScrapeResponse{}
	for contents := range pageContentsCh {
		pageContents = append(pageContents, contents)
	}

	return pageContents, nil
}
