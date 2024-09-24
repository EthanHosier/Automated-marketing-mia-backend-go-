package campaign_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ethanhosier/mia-backend-go/canva"
	"github.com/ethanhosier/mia-backend-go/images"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

const (
	retryAttempts = 3
)

type CampaignHelper interface {
	GetCandidatePageContentsForUser(userID string, n int) ([]researcher.PageContents, error)
	GenerateThemes(pageContents []researcher.PageContents, businessSummary *researcher.BusinessSummary) ([]CampaignTheme, error)
	TemplatePlan(templatePrompt string, templateToFill storage.Template) (*ExtractedTemplate, error)
	InitFields(ctxt context.Context, template *ExtractedTemplate, campaignDetailsStr string, candidateImages []string) ([]canva.TextField, []canva.ImageField, []canva.ColorField, error)
}

type CampaignHelperClient struct {
	openaiClient openai.OpenaiClient
	researcher   researcher.Researcher
	canvaClient  canva.CanvaClient
	storage      storage.Storage
	imagesClient images.ImagesClient
}

func NewCampaignHelperClient(openaiClient openai.OpenaiClient, researcher researcher.Researcher, canvaClient canva.CanvaClient, storage storage.Storage, imagesClient images.ImagesClient) *CampaignHelperClient {
	return &CampaignHelperClient{
		openaiClient: openaiClient,
		researcher:   researcher,
		canvaClient:  canvaClient,
		storage:      storage,
		imagesClient: imagesClient,
	}
}

func (c *CampaignHelperClient) TemplatePlan(templatePrompt string, templateToFill storage.Template) (*ExtractedTemplate, error) {
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

	return c.TemplateWithCorrectedTextFields(&extractedTemplate, templateToFill)
}

func (c *CampaignHelperClient) TemplateWithCorrectedTextFields(extractedTemplate *ExtractedTemplate, templateToFill storage.Template) (*ExtractedTemplate, error) {
	maxCharMap := map[string]int{}
	for _, field := range templateToFill.Fields {
		if field.Type == "text" {
			maxCharMap[field.Name] = field.MaxCharacters
		}
	}

	textFields, imgFields := []PopulatedField{}, []PopulatedField{}
	textUpdateTasks := []*utils.Task[*PopulatedField]{}

	for _, field := range extractedTemplate.Fields {
		if field.Type == TextType {
			if maxChars, ok := maxCharMap[field.Name]; ok && len(field.Value) > maxChars {
				textUpdateTasks = append(textUpdateTasks, utils.DoAsync[*PopulatedField](func() (*PopulatedField, error) {
					return c.rephraseTextFieldCharsWithRetry(maxChars, &field)
				}))
			} else {
				textFields = append(textFields, field)
			}
		} else if field.Type == ImageType {
			imgFields = append(imgFields, field)
		}
	}

	for _, task := range textUpdateTasks {
		textField, err := utils.GetAsync(task)
		if err != nil {
			return nil, err
		}

		textFields = append(textFields, *textField)
	}

	return &ExtractedTemplate{
		Platform:    extractedTemplate.Platform,
		Caption:     extractedTemplate.Caption,
		Fields:      append(textFields, imgFields...),
		ColorFields: extractedTemplate.ColorFields,
	}, nil
}

func (c *CampaignHelperClient) rephraseTextFieldCharsWithRetry(maxChars int, field *PopulatedField) (*PopulatedField, error) {
	t, err := utils.Retry[string](3, func() (string, error) {
		rephrased, err := c.openaiClient.ChatCompletion(context.TODO(), fmt.Sprintf(openai.MaxCharsPrompt, maxChars, field.Value), openai.GPT4o)

		if len(rephrased) > maxChars {
			slog.Warn("Rephrased text too long", "rephrased", rephrased)
			return rephrased[:maxChars], nil
		}

		return rephrased, err
	})
	if err != nil {
		return nil, err
	}

	return &PopulatedField{
		Name:  field.Name,
		Type:  field.Type,
		Value: t,
	}, err
}

func (c *CampaignHelperClient) InitFields(ctxt context.Context, template *ExtractedTemplate, campaignDetailsStr string, candidateImages []string) ([]canva.TextField, []canva.ImageField, []canva.ColorField, error) {
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
		return c.initImageFields(ctxt, imageUploadFields, candidateImages, campaignDetailsStr)
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

func (c *CampaignHelperClient) initColorFields(colorUploadFields []PopulatedColorField) ([]canva.ColorField, error) {
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

func (c *CampaignHelperClient) initImageFields(ctxt context.Context, imageUploadFields []PopulatedField, candidateImages []string, campaignDetailsStr string) ([]canva.ImageField, error) {

	bestImages, err := c.bestImages(ctxt, imageUploadFields, candidateImages, campaignDetailsStr)
	if err != nil {
		return nil, err
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

func (c *CampaignHelperClient) bestImages(ctxt context.Context, imageUploadFields []PopulatedField, candidateImages []string, campaignDetailsStr string) ([]string, error) {
	imageSet := make(map[string]struct{})
	mu := sync.Mutex{}

	bestImageTasks := utils.DoAsyncList(imageUploadFields, func(p PopulatedField) (string, error) {
		var (
			maxRetries = 5
			attempts   = 0
			img        string
			err        error
			imgs       = candidateImages
		)

		captions, err := c.getCaptionsCompletionArr(p.Value)
		if err != nil {
			return "", err
		}

		for attempts < maxRetries {
			attempts++

			img, err = c.imagesClient.BestImageFor(ctxt, captions, imgs, campaignDetailsStr, p.Value)
			if err != nil {
				return "", err
			}

			mu.Lock()

			if _, ok := imageSet[img]; ok {
				imgs = utils.RemoveElements(imgs, utils.GetKeysFromMap(imageSet))

				mu.Unlock()
				slog.Info("Duplicate image found, retrying", "image", img)
				continue
			}

			imageSet[img] = struct{}{}
			mu.Unlock()

			return img, nil
		}

		return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
	})

	return utils.GetAsyncList(bestImageTasks)
}

func (c *CampaignHelperClient) GetCandidatePageContentsForUser(userID string, n int) ([]researcher.PageContents, error) {
	randomUrls, err := storage.GetRandom[researcher.SitemapUrl](c.storage, n, map[string]string{"id": userID})
	if err != nil {
		return nil, err
	}

	slog.Info("Random URLs", "urls", randomUrls)

	pageContentsTasks := []*utils.Task[*researcher.PageContents]{}

	for _, url := range randomUrls {
		pageContentsTasks = append(pageContentsTasks, utils.DoAsync[*researcher.PageContents](func() (*researcher.PageContents, error) {
			return c.researcher.PageContentsFor(url.Url)
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

func (c *CampaignHelperClient) GenerateThemes(pageContents []researcher.PageContents, businessSummary *researcher.BusinessSummary) ([]CampaignTheme, error) {
	themePrompt := fmt.Sprintf(openai.ThemeGenerationPrompt, businessSummary, pageContents, businessSummary.TargetRegion, "", "")

	themesWithSuggestedKeywords, err := utils.Retry(retryAttempts, func() ([]themeWithSuggestedKeywords, error) {
		return c.themes(themePrompt)
	})

	if err != nil {
		return nil, err
	}

	return c.themesWithChosenKeywords(themesWithSuggestedKeywords)
}

func (c *CampaignHelperClient) themesWithChosenKeywords(themesWithSuggestedKeywords []themeWithSuggestedKeywords) ([]CampaignTheme, error) {

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

func (c *CampaignHelperClient) chosenKeywords(keywords []string) (string, string, error) {
	adsKeywords, err := c.researcher.GoogleAdsKeywordsData(keywords)

	if err != nil {
		return "", "", fmt.Errorf("error getting Google Ads data: %w", err)
	}

	return c.researcher.OptimalKeywords(adsKeywords)
}

func (c *CampaignHelperClient) themes(themePrompt string) ([]themeWithSuggestedKeywords, error) {
	completion, err := c.openaiClient.ChatCompletion(context.TODO(), themePrompt, openai.GPT4oMini)

	if err != nil {
		return nil, err
	}

	extractedArr := openai.ExtractJsonData(completion, openai.JSONArray)

	var themes []themeWithSuggestedKeywords
	err = json.Unmarshal([]byte(extractedArr), &themes)
	if err != nil {
		return nil, err
	}

	return themes, nil
}

//       2. Add max character prompt
