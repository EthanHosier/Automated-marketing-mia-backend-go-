package campaigns

import (
	"fmt"
	"testing"

	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestChosenKeywords(t *testing.T) {
	// given
	var (
		r = researcher.NewMockResearcher()

		c = NewCampaignClient(nil, r, nil, nil)

		keywords    = []string{"keyword1"}
		adsKeywords = []researcher.GoogleAdsKeyword{
			{
				Keyword:            "keyword1",
				AvgMonthlySearches: 100,
				CompetitionLevel:   "low",
				CompetitionIndex:   1,
				LowTopOfPageBid:    1,
				HighTopOfPageBid:   2,
			},
		}
	)

	r.GoogleAdsKeywordsDataWillReturn(keywords, adsKeywords, nil)
	r.OptimalKeywordsWillReturn(adsKeywords, "prim", "sec", nil)

	// when
	primaryKeyword, secondaryKeyword, err := c.chosenKeywords(keywords)

	// then
	assert.NoError(t, err)
	assert.Equal(t, "prim", primaryKeyword)
	assert.Equal(t, "sec", secondaryKeyword)
}

func TestThemesWithGivenKeywords(t *testing.T) {
	// given
	var (
		r = researcher.NewMockResearcher()
		c = NewCampaignClient(nil, r, nil, nil)

		keywords    = []string{"keyword1"}
		adsKeywords = []researcher.GoogleAdsKeyword{
			{
				Keyword:            "keyword1",
				AvgMonthlySearches: 100,
				CompetitionLevel:   "low",
				CompetitionIndex:   1,
				LowTopOfPageBid:    1,
				HighTopOfPageBid:   2,
			},
		}

		themesWithSuggestedKeywords = []themeWithSuggestedKeywords{{
			Keywords: keywords,
		}}

		campaignThemes = []CampaignTheme{{
			PrimaryKeyword:   "prim",
			SecondaryKeyword: "sec",
		}}
	)

	// given
	r.GoogleAdsKeywordsDataWillReturn(keywords, adsKeywords, nil)
	r.OptimalKeywordsWillReturn(adsKeywords, "prim", "sec", nil)

	// when
	res, err := c.themesWithChosenKeywords(themesWithSuggestedKeywords)

	// then
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, campaignThemes[0].PrimaryKeyword, res[0].PrimaryKeyword)
	assert.Equal(t, campaignThemes[0].SecondaryKeyword, res[0].SecondaryKeyword)
}

func TestCandidatePagesForUser(t *testing.T) {
	// given
	var (
		r = researcher.NewMockResearcher()
		s = storage.NewInMemoryStorage()
		c = NewCampaignClient(nil, r, nil, s)

		userID     = "user1"
		sitemapUrl = researcher.SitemapUrl{
			ID:           "id1",
			Url:          "url1",
			UrlEmbedding: []float32{1.0},
		}

		pageContents = []researcher.PageContents{
			{
				Url: "url1",
			},
		}
	)

	r.PageContentsForWillReturn(sitemapUrl.Url, &pageContents[0], nil)

	// when
	storage.Store(s, sitemapUrl)
	res, err := c.getCandidatePageContentsForUser(userID, 1)

	// then
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, pageContents[0].Url, res[0].Url)
}

func TestBestImage(t *testing.T) {
	// given
	var (
		op = openai.MockOpenaiClient{}
		c  = NewCampaignClient(&op, nil, nil, nil)

		campaignDetailsStr = "campaignDetails"
		imageDescription   = "imageDescription"
		prompt             = fmt.Sprintf(openai.PickBestImagePrompt, campaignDetailsStr, imageDescription)
		images             = []string{"image1", "image2"}
	)

	op.WillReturnImageCompletion(prompt, images, openai.GPT4o, "1")

	// when
	res, err := c.bestImage(imageDescription, images, campaignDetailsStr)

	// then
	assert.NoError(t, err)
	assert.Equal(t, images[1], res)
}

func TestThemes(t *testing.T) {
	// given
	var (
		op = openai.MockOpenaiClient{}
		c  = NewCampaignClient(&op, nil, nil, nil)

		themePrompt = "themePrompt"
		theme1      = themeWithSuggestedKeywords{
			Theme:                         "Modern",
			Keywords:                      []string{"sleek", "contemporary", "minimal"},
			Url:                           "https://example.com/modern",
			SelectedUrl:                   "https://example.com/modern/selected",
			ImageCanvaTemplateDescription: "A modern and sleek design template.",
		}

		theme2 = themeWithSuggestedKeywords{
			Theme:                         "Vintage",
			Keywords:                      []string{"retro", "classic", "timeless"},
			Url:                           "https://example.com/vintage",
			SelectedUrl:                   "https://example.com/vintage/selected",
			ImageCanvaTemplateDescription: "A vintage and classic design template.",
		}

		themesStr = `[
				{
						"theme": "Modern",
						"keywords": ["sleek", "contemporary", "minimal"],
						"url": "https://example.com/modern",
						"selectedUrl": "https://example.com/modern/selected",
						"imageCanvaTemplateDescription": "A modern and sleek design template."
				},
				{
						"theme": "Vintage",
						"keywords": ["retro", "classic", "timeless"],
						"url": "https://example.com/vintage",
						"selectedUrl": "https://example.com/vintage/selected",
						"imageCanvaTemplateDescription": "A vintage and classic design template."
				}
		]`
	)

	op.WillReturnChatCompletion(themePrompt, openai.GPT4oMini, themesStr)

	// when
	res, err := c.themes(themePrompt)

	// then
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, theme1, res[0])
	assert.Equal(t, theme2, res[1])
}
