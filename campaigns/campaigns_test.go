package campaigns

import (
	"testing"

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
