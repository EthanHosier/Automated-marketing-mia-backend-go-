package campaigns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethanhosier/mia-backend-go/canva"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

const (
	numberOfThemes                  = 5
	retryAttempts                   = 3
	maxScrapedPageBodyTextCharCount = 4000
)

type CampaignClient struct {
	openaiClient *openai.OpenaiClient
	researcher   *researcher.Researcher
	canvaClient  *canva.CanvaClient
	storage      *storage.Storage
}

func NewCampaignClient(openaiClient *openai.OpenaiClient, researcher *researcher.Researcher, canvaClient *canva.CanvaClient, storage *storage.Storage) *CampaignClient {
	return &CampaignClient{
		openaiClient: openaiClient,
		researcher:   researcher,
		storage:      storage,
		canvaClient:  canvaClient,
	}
}

func (c *CampaignClient) GenerateThemesForUser(userID string) ([]CampaignTheme, error) {
	candidatePageContents, err := c.getCandidatePageContentsForUser(userID, numberOfThemes)
	if err != nil {
		return nil, err
	}

	businessSummary, err := storage.Get[researcher.BusinessSummary](c.storage, userID)
	if err != nil {
		return nil, err
	}

	return c.generateThemes(candidatePageContents, businessSummary)
}

func (c *CampaignClient) CampaignFrom(theme CampaignTheme, businessSummary *researcher.BusinessSummary) (*canva.Design, *string, error) {

	scrapedPageBodyTask := utils.DoAsync[string](func() (string, error) {
		return c.researcher.PageBodyTextFor(theme.Url)
	})

	scrapedPageContentsTask := utils.DoAsync[*researcher.PageContents](func() (*researcher.PageContents, error) {
		return c.researcher.PageContentsFor(theme.Url)
	})

	posts, err := c.researcher.SocialMediaPostsFor(theme.PrimaryKeyword)
	if err != nil {
		return nil, nil, err
	}

	researchReportTask := utils.DoAsync[string](func() (string, error) {
		return c.researcher.ResearchReportFromPosts(posts)
	})

	templates, err := storage.GetRandom[storage.Template](c.storage, 1)
	if err != nil {
		return nil, nil, err
	}

	scrapedPageBodyText, err := utils.GetAsync(scrapedPageBodyTask)
	if err != nil {
		return nil, nil, err
	}

	templatePrompt := templatePrompt(
		researcher.Facebook,
		*businessSummary,
		theme.Theme,
		theme.PrimaryKeyword,
		theme.SecondaryKeyword,
		theme.Url,
		scrapedPageBodyText,
		posts,
		templates[0].Fields,
		templates[0].ColorFields,
	)

	templatePlan, err := c.templatePlan(templatePrompt)
	if err != nil {
		return nil, nil, err
	}

	scrapedPageContents, err := utils.GetAsync(scrapedPageContentsTask)
	if err != nil {
		return nil, nil, err
	}

	campaignDetailsStr := fmt.Sprintf("Primary keyword: %v\nSecondary keyword: %v\nURL: %v\nTheme: %v\nTemplate Description: %v", theme.PrimaryKeyword, theme.SecondaryKeyword, theme.Url, theme.Theme, theme.ImageCanvaTemplateDescription)

	textFields, imageFields, colorFields, err := c.initFields(templatePlan, campaignDetailsStr, scrapedPageContents.ImageUrls)
	if err != nil {
		return nil, nil, err
	}

	template, err := c.canvaClient.PopulateTemplate(templates[0].ID, imageFields, textFields, colorFields)
	if err != nil {
		return nil, nil, err
	}

	researchReport, err := utils.GetAsync(researchReportTask)
	if err != nil {
		return nil, nil, err
	}

	return &template.Design, &researchReport, nil
}

func (c *CampaignClient) templatePlan(templatePrompt string) (*ExtractedTemplate, error) {
	templateCompletion, err := c.openaiClient.ChatCompletion(context.TODO(), templatePrompt, openai.GPT4o)
	if err != nil {
		return nil, err
	}

	jsonStr := openai.ExtractJsonData(templateCompletion, openai.JSONObj)
	var extractedTemplate ExtractedTemplate
	err = json.Unmarshal([]byte(jsonStr), &extractedTemplate)
	if err != nil {
		return nil, err
	}

	return &extractedTemplate, nil
}

func (c *CampaignClient) initFields(template *ExtractedTemplate, campaignDetailsStr string, candidateImages []string) ([]canva.TextField, []canva.ImageField, []canva.ColorField, error) {
	textFields := []canva.TextField{}

	imageUploadFields := []PopulatedField{}

	for _, field := range template.Fields {
		if field.Type == TextType {
			textFields = append(textFields, canva.TextField{
				Name: field.Name,
				Text: field.Value,
			})
		} else if field.Type == ImageType {
			imageUploadFields = append(imageUploadFields, field)
		}
	}

	imageFieldsTask := utils.DoAsync[[]canva.ImageField](func() ([]canva.ImageField, error) {
		return c.initImageFields(imageUploadFields, candidateImages, campaignDetailsStr)
	})

	colorFieldsTask := utils.DoAsync[[]canva.ColorField](func() ([]canva.ColorField, error) {
		return c.initColorFields(template.ColorFields)
	})

	imageFields, err := utils.GetAsync(imageFieldsTask)
	if err != nil {
		return nil, nil, nil, err
	}

	colorFields, err := utils.GetAsync(colorFieldsTask)
	if err != nil {
		return nil, nil, nil, err
	}

	return textFields, imageFields, colorFields, nil
}

func (c *CampaignClient) initColorFields(colorUploadFields []PopulatedColorField) ([]canva.ColorField, error) {
	colors := []string{}
	for _, field := range colorUploadFields {
		colors = append(colors, field.Color)
	}

	assetIds, err := c.canvaClient.UploadColorAssets(colors)
	if err != nil {
		return nil, err
	}

	colorFields := []canva.ColorField{}
	for i, field := range colorUploadFields {
		colorFields = append(colorFields, canva.ColorField{
			Name:         field.Name,
			ColorAssetId: assetIds[i],
		})
	}

	return colorFields, nil
}

func (c *CampaignClient) initImageFields(imageUploadFields []PopulatedField, candidateImages []string, campaignDetailsStr string) ([]canva.ImageField, error) {
	bestImageTasks := []*utils.Task[string]{}

	for _, field := range imageUploadFields {
		bestImageTasks = append(bestImageTasks, utils.DoAsync[string](func() (string, error) {
			return c.bestImage(field.Value, candidateImages, campaignDetailsStr)
		}))
	}

	bestImages := []string{}
	for _, task := range bestImageTasks {
		bestImage, err := utils.GetAsync(task)
		if err != nil {
			return nil, err
		}
		bestImages = append(bestImages, bestImage)
	}

	assetIds, err := c.canvaClient.UploadImageAssets(bestImages)
	if err != nil {
		return nil, err
	}

	imageFields := []canva.ImageField{}

	for i, field := range imageUploadFields {
		imageFields = append(imageFields, canva.ImageField{
			Name:    field.Name,
			AssetId: assetIds[i],
		})
	}

	return imageFields, nil
}

func (c *CampaignClient) bestImage(imageDescription string, candidateImages []string, campaignDetailsStr string) (string, error) {
	if len(candidateImages) > 50 {
		return "", fmt.Errorf("> 50 candidate images")
	}

	if len(candidateImages) == 0 {
		return "", fmt.Errorf("no candidate images supplied")
	}

	prompt := fmt.Sprint(openai.PickBestImagePrompt, campaignDetailsStr, imageDescription)
	bestImage, err := c.openaiClient.ImageCompletion(context.TODO(), prompt, candidateImages, openai.GPT4o)
	if err != nil {
		return "", err
	}

	i, err := utils.FirstNumberInString(bestImage)
	if err != nil {
		return "", err
	}

	return candidateImages[i], nil
}

func (c *CampaignClient) getCandidatePageContentsForUser(userID string, n int) ([]researcher.PageContents, error) {
	randomUrls, err := storage.GetRandom[researcher.Sitemap](c.storage, n)
	if err != nil {
		return nil, err
	}

	pageContentsTasks := []*utils.Task[*researcher.PageContents]{}

	for _, url := range randomUrls {
		pageContentsTasks = append(pageContentsTasks, utils.DoAsync[*researcher.PageContents](func() (*researcher.PageContents, error) {
			return c.researcher.PageContentsFor(url)
		}))
	}

	pageContents := []researcher.PageContents{}
	for _, task := range pageContentsTasks {
		pageContent, err := utils.GetAsync(task)
		if err != nil {
			return nil, err
		}
		pageContents = append(pageContents, *pageContent)
	}

	return pageContents, nil
}

func (c *CampaignClient) generateThemes(pageContents []researcher.PageContents, businessSummary *researcher.BusinessSummary) ([]CampaignTheme, error) {
	themePrompt := fmt.Sprintf(openai.ThemeGenerationPrompt, businessSummary, pageContents, businessSummary.TargetRegion, "", "")

	themesWithSuggestedKeywords, err := utils.Retry(retryAttempts, func() ([]themeWithSuggestedKeywords, error) {
		return c.themes(themePrompt)
	})

	if err != nil {
		return nil, err
	}

	return c.themesWithChosenKeywords(themesWithSuggestedKeywords)
}

func (c *CampaignClient) themesWithChosenKeywords(themesWithSuggestedKeywords []themeWithSuggestedKeywords) ([]CampaignTheme, error) {

	campaignThemesTasks := []*utils.Task[*CampaignTheme]{}
	for _, t := range themesWithSuggestedKeywords {
		campaignThemesTasks = append(campaignThemesTasks, utils.DoAsync[*CampaignTheme](func() (*CampaignTheme, error) {
			primaryKeyword, secondaryKeyword, err := c.chosenKeywords(t.Keywords)
			if err != nil {
				return nil, err
			}

			return &CampaignTheme{
				Theme:                         t.Theme,
				PrimaryKeyword:                primaryKeyword,
				SecondaryKeyword:              secondaryKeyword,
				Url:                           t.Url,
				SelectedUrl:                   t.SelectedUrl,
				ImageCanvaTemplateDescription: t.ImageCanvaTemplateDescription,
			}, nil
		}))
	}

	campaignThemes := []CampaignTheme{}
	for _, ct := range campaignThemesTasks {
		campaignTheme, err := utils.GetAsync(ct)
		if err != nil {
			return nil, err
		}
		campaignThemes = append(campaignThemes, *campaignTheme)
	}

	return campaignThemes, nil
}

func (c *CampaignClient) chosenKeywords(keywords []string) (string, string, error) {
	adsKeywords, err := c.researcher.GoogleAdsKeywordsData(keywords)

	if err != nil {
		return "", "", fmt.Errorf("error getting Google Ads data: %w", err)
	}

	return c.researcher.OptimalKeywords(adsKeywords)
}
