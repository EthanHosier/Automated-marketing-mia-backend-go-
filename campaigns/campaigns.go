package campaigns

import (
	"context"
	"fmt"

	"github.com/ethanhosier/mia-backend-go/campaigns/campaign_helper"
	"github.com/ethanhosier/mia-backend-go/canva"
	"github.com/ethanhosier/mia-backend-go/images"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

const (
	numberOfThemes = 5
)

type CampaignClient struct {
	campaignHelper campaign_helper.CampaignHelper
	storage        storage.Storage
	researcher     researcher.Researcher
	canvaClient    canva.CanvaClient
	imagesClient   images.ImagesClient
}

func NewCampaignClient(openaiClient openai.OpenaiClient, researcher researcher.Researcher, canvaClient canva.CanvaClient, storage storage.Storage, imagesClient images.ImagesClient, campaignHelper campaign_helper.CampaignHelper) *CampaignClient {
	return &CampaignClient{
		campaignHelper: campaignHelper,
		storage:        storage,
		researcher:     researcher,
		canvaClient:    canvaClient,
		imagesClient:   imagesClient,
	}
}

func (c *CampaignClient) GenerateThemesForUser(userID string) ([]campaign_helper.CampaignTheme, error) {
	candidatePageContents, err := c.campaignHelper.GetCandidatePageContentsForUser(userID, numberOfThemes)
	if err != nil {
		return nil, err
	}

	businessSummary, err := storage.Get[researcher.BusinessSummary](c.storage, userID)
	if err != nil {
		return nil, err
	}

	return c.campaignHelper.GenerateThemes(candidatePageContents, businessSummary)
}

func (c *CampaignClient) CampaignFrom(ctxt context.Context, theme campaign_helper.CampaignTheme, businessSummary *researcher.BusinessSummary) ([]*storage.Post, string, error) {

	scrapedPageBodyTask := utils.DoAsync[string](func() (string, error) {
		return c.researcher.PageBodyTextFor(theme.Url)
	})

	scrapedPageContentsTask := utils.DoAsync[*researcher.PageContents](func() (*researcher.PageContents, error) {
		return c.researcher.PageContentsFor(theme.Url)
	})

	posts, err := c.researcher.SocialMediaPostsFor(theme.PrimaryKeyword)
	if err != nil {
		return nil, "", err
	}

	researchReportTask := utils.DoAsync[string](func() (string, error) {
		return c.researcher.ResearchReportFromPosts(posts)
	})

	templates, err := storage.GetRandom[storage.Template](c.storage, len(researcher.SocialMediaPlatforms), nil)
	if err != nil {
		return nil, "", err
	}

	scrapedPageBodyText, err := utils.GetAsync(scrapedPageBodyTask)
	if err != nil {
		return nil, "", err
	}

	scrapedPageContents, err := utils.GetAsync(scrapedPageContentsTask)
	if err != nil {
		return nil, "", err
	}

	campaignDetailsStr := fmt.Sprintf("Primary keyword: %v\nSecondary keyword: %v\nURL: %v\nTheme: %v\nTemplate Description: %v", theme.PrimaryKeyword, theme.SecondaryKeyword, theme.Url, theme.Theme, theme.ImageCanvaTemplateDescription)

	tasks := []*utils.Task[*storage.Post]{}
	for i, template := range templates {
		templatePrompt := templatePrompt(
			researcher.SocialMediaPlatforms[i],
			*businessSummary,
			theme.Theme,
			theme.PrimaryKeyword,
			theme.SecondaryKeyword,
			theme.Url,
			scrapedPageBodyText,
			posts,
			template.Fields,
			template.ColorFields,
		)

		tasks = append(tasks, utils.DoAsync(func() (*storage.Post, error) {
			return c.templateFrom(ctxt, templatePrompt, campaignDetailsStr, *scrapedPageContents, template, researcher.SocialMediaPlatforms[i])
		}))
	}

	researchReport, err := utils.GetAsync(researchReportTask)
	if err != nil {
		return nil, "", err
	}

	postResponses, err := utils.GetAsyncList(tasks)

	return postResponses, researchReport, err
}

func (c *CampaignClient) templateFrom(ctxt context.Context, templatePrompt string, campaignDetailsStr string, scrapedPageContents researcher.PageContents, template storage.Template, platform researcher.SocialMediaPlatform) (*storage.Post, error) {
	templatePlan, err := c.campaignHelper.TemplatePlan(templatePrompt, template)
	fmt.Printf("Template Plan: %+v\n\n", templatePlan)
	if err != nil {
		return nil, err
	}

	candidateImages, err := c.imagesClient.FilterTooSmallImages(scrapedPageContents.ImageUrls)
	if err != nil {
		return nil, err
	}

	textFields, imageFields, colorFields, err := c.campaignHelper.InitFields(ctxt, templatePlan, campaignDetailsStr, candidateImages)
	if err != nil {
		return nil, err
	}

	canvaResult, err := c.canvaClient.PopulateTemplate(template.ID, imageFields, textFields, colorFields)

	if err != nil {
		return nil, err
	}

	postResponse := &storage.Post{
		Platform: string(platform),
		Caption:  templatePlan.Caption,
		Design:   canvaResult.Design,
	}

	return postResponse, nil
}
