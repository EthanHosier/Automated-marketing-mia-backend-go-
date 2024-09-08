package campaigns

import (
	"fmt"
	"sync"

	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

const (
	numberOfThemes = 5
	retryAttempts  = 3
)

type CampaignClient struct {
	openaiClient *openai.OpenaiClient
	researcher   *researcher.Researcher
	storage      *storage.Storage
}

func NewCampaignClient(openaiClient *openai.OpenaiClient, researcher *researcher.Researcher, storage *storage.Storage) *CampaignClient {
	return &CampaignClient{
		openaiClient: openaiClient,
		researcher:   researcher,
		storage:      storage,
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

func (c *CampaignClient) getCandidatePageContentsForUser(userID string, n int) ([]researcher.PageContents, error) {
	randomUrls, err := storage.GetRandom[researcher.Sitemap](c.storage, n)
	if err != nil {
		return nil, err
	}

	pageContentsCh := make(chan *researcher.PageContents, n)
	errorCh := make(chan error, n)

	var pageContentsWg sync.WaitGroup
	pageContentsWg.Add(n)

	for _, url := range randomUrls {
		go func(url string) {
			defer pageContentsWg.Done()
			pageContents, err := c.researcher.PageContentsFor(url)
			if err != nil {
				errorCh <- err
				return
			}
			pageContentsCh <- pageContents
		}(url)
	}

	pageContentsWg.Wait()
	close(pageContentsCh)

	select {
	case err := <-errorCh:
		return nil, err

	default:
	}

	pageContents := []researcher.PageContents{}
	for pc := range pageContentsCh {
		pageContents = append(pageContents, *pc)
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

	campaignThemeCh := make(chan *CampaignTheme, len(themesWithSuggestedKeywords))
	errorCh := make(chan error, len(themesWithSuggestedKeywords))

	themeWithKeywordsWg := sync.WaitGroup{}
	themeWithKeywordsWg.Add(len(themesWithSuggestedKeywords))

	for _, t := range themesWithSuggestedKeywords {
		go func(t themeWithSuggestedKeywords) {
			defer themeWithKeywordsWg.Done()

			primaryKeyword, secondaryKeyword, err := c.chosenKeywords(t.Keywords)
			if err != nil {
				errorCh <- err
				return
			}

			campaignThemeCh <- &CampaignTheme{
				Theme:                         t.Theme,
				PrimaryKeyword:                primaryKeyword,
				SecondaryKeyword:              secondaryKeyword,
				Url:                           t.Url,
				SelectedUrl:                   t.SelectedUrl,
				ImageCanvaTemplateDescription: t.ImageCanvaTemplateDescription,
			}
		}(t)
	}

	themeWithKeywordsWg.Wait()
	close(campaignThemeCh)
	close(errorCh)

	select {
	case err := <-errorCh:
		return nil, err
	default:
	}

	campaignThemes := []CampaignTheme{}
	for ct := range campaignThemeCh {
		campaignThemes = append(campaignThemes, *ct)
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
