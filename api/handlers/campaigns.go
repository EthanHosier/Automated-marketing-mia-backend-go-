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

		log.Println("generated campaign", campaign)
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

func campaignFromTheme(theme types.ThemeData, businessSummary types.StoredBusinessSummary, llmClient *utils.LLMClient, store storage.Storage) (string, error) {
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
	templateCh := make(chan *types.NearestTemplateResponse, 1)
	go func() {
		template, err := chosenTemplate(theme.ImageCanvaTemplateDescription, llmClient, store)
		if err != nil {
			fmt.Errorf("error getting template for campaign: %w", err)
			templateCh <- nil
			return
		}
		templateCh <- template
	}()

	primaryKeyword, secondaryKeyword, err := campaignChosenKewords(theme.Keywords)
	if err != nil {
		return "", fmt.Errorf("error getting keywords for campaign: %w", err)
	}

	researchReportData, err := utils.ResearchReportData(primaryKeyword, llmClient)
	if err != nil {
		return "", fmt.Errorf("error getting research report data: %w", err)
	}

	// Generate Research Report
	researchReportCh := make(chan string, 1)
	go func() {
		researchReportPrompt := prompts.ResearchReportPrompt(primaryKeyword, researchReportData)
		researchReport, err := llmClient.OpenaiCompletion(researchReportPrompt, openai.GPT4o)
		if err != nil {
			log.Println("error generating research report: %w", err)
			researchReportCh <- ""
			return
		}
		researchReportCh <- researchReport
	}()

	// Gather the template to be used for the campaign
	template := <-templateCh
	scrapedPageBody := <-scrapedPageBodyCh

	templatePrompt := prompts.TemplatePrompt("instagram", businessSummary, theme.Theme, primaryKeyword, secondaryKeyword, url, scrapedPageBody,
		researchReportData, template.Fields, template.ColorFields)

	templateCompletion, err := llmClient.OpenaiCompletion(templatePrompt, openai.GPT4o)
	if err != nil {
		return "", fmt.Errorf("error generating template completion: %w", err)
	}
	extractedTemplate := utils.ExtractJsonObj(templateCompletion, utils.CurlyBracket)

	var populatedTemplate types.PopulatedTemplate
	err = json.Unmarshal([]byte(extractedTemplate), &populatedTemplate)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling populated template: %w.\n Template completion was %+v\n Extracted template was %+v", err, templateCompletion, extractedTemplate)
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

	pageContents := <-scrapedPageContentsCh
	bestImages, err := utils.PickBestImages(pageContents.ImageUrls, campaignDetailsStr, imageFields, llmClient)
	if err != nil {
		return "", fmt.Errorf("error picking best images: %w", err)
	}

	imageFields, err = uploadImageAssets(imageFields, bestImages)
	if err != nil {
		return "", fmt.Errorf("error uploading image assets: %w", err)
	}

	colorFields, err := uploadColorAssets(populatedTemplate.ColorFields)
	if err != nil {
		return "", fmt.Errorf("error uploading color assets: %w", err)
	}

	log.Printf("Populated template: %+v\n", populatedTemplate)

	err = utils.PopulateTemplate(template.ID, imageFields, textFields, colorFields)
	if err != nil {
		return "", fmt.Errorf("error populating template: %w", err)
	}

	return "", nil
}

func uploadColorAssets(colorFields []types.PopulatedColorField) ([]types.PopulatedColorField, error) {
	colorFieldCh := make(chan types.PopulatedColorField, len(colorFields))
	colorFieldWg := sync.WaitGroup{}
	colorFieldWg.Add(len(colorFields))

	for _, field := range colorFields {
		go func(field types.PopulatedColorField) {
			defer colorFieldWg.Done()

			colorImg, err := utils.CreateColorImage(field.Color)
			if err != nil {
				log.Println("error creating color image: ", err)
				return
			}

			resp, err := utils.UploadAsset(colorImg, "name")
			if err != nil {
				log.Println("error uploading color image: ", err)
				return
			}

			field.Color = resp.Job.ID
			colorFieldCh <- field
		}(field)
	}

	colorFieldWg.Wait()
	close(colorFieldCh)

	colorFields = []types.PopulatedColorField{}
	for field := range colorFieldCh {
		colorFields = append(colorFields, field)
	}

	return colorFields, nil
}

func uploadImageAssets(imageFields []types.PopulatedField, bestImages []string) ([]types.PopulatedField, error) {
	imageFieldsCh := make(chan types.PopulatedField, len(imageFields))
	imageFieldsWg := sync.WaitGroup{}
	imageFieldsWg.Add(len(imageFields))

	for i, field := range imageFields {
		go func(field types.PopulatedField, i int) {
			defer imageFieldsWg.Done()
			img, err := utils.DownloadImage(bestImages[i])
			if err != nil {
				log.Println("error downloading image: ", err)
				return
			}

			resp, err := utils.UploadAsset(img, "name")
			if err != nil {
				log.Println("error uploading image: ", err)
				return
			}

			field.Value = resp.Job.ID
			imageFieldsCh <- field
		}(field, i)
	}

	imageFieldsWg.Wait()
	close(imageFieldsCh)

	imageFields = []types.PopulatedField{}
	for field := range imageFieldsCh {
		imageFields = append(imageFields, field)
	}

	return imageFields, nil
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
	// embedding, err := llmClient.OpenaiEmbedding(templateDescription)
	// if err != nil {
	// 	return nil, fmt.Errorf("error getting embedding for optimal keyword: %w", err)
	// }

	// template, err := store.GetNearestTemplate(embedding)
	template, err := store.GetRandomTemplate()
	if err != nil {
		return nil, fmt.Errorf("error getting nearest template: %w", err)
	}

	return template, nil
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
