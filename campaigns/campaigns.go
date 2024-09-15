package campaigns

import (
	"fmt"

	"github.com/ethanhosier/mia-backend-go/campaigns/campaign_helper"
	"github.com/ethanhosier/mia-backend-go/canva"
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
}

func NewCampaignClient(openaiClient openai.OpenaiClient, researcher researcher.Researcher, canvaClient canva.CanvaClient, storage storage.Storage) *CampaignClient {
	ch := campaign_helper.NewCampaignHelperClient(openaiClient, researcher, canvaClient, storage)

	return &CampaignClient{
		campaignHelper: ch,
		storage:        storage,
		researcher:     researcher,
		canvaClient:    canvaClient,
	}
}

// TODO: move everything else other than this and the below to another file (campaign helper)
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

// TODO: move everything else other than this and the above to another file (campaign helper)
func (c *CampaignClient) CampaignFrom(theme campaign_helper.CampaignTheme, businessSummary *researcher.BusinessSummary) (*canva.Design, *string, error) {

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

	templatePlan, err := c.campaignHelper.TemplatePlan(templatePrompt)
	if err != nil {
		return nil, nil, err
	}

	scrapedPageContents, err := utils.GetAsync(scrapedPageContentsTask)
	if err != nil {
		return nil, nil, err
	}

	campaignDetailsStr := fmt.Sprintf("Primary keyword: %v\nSecondary keyword: %v\nURL: %v\nTheme: %v\nTemplate Description: %v", theme.PrimaryKeyword, theme.SecondaryKeyword, theme.Url, theme.Theme, theme.ImageCanvaTemplateDescription)

	textFields, imageFields, colorFields, err := c.campaignHelper.InitFields(templatePlan, campaignDetailsStr, scrapedPageContents.ImageUrls)
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
