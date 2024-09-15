package campaigns

import (
	"testing"

	"github.com/ethanhosier/mia-backend-go/campaigns/campaign_helper"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestGenerateThemesForUser(t *testing.T) {
	// given
	var (
		mockCampaignHelper = campaign_helper.NewMockCampaignHelper()
		mockStorage        = storage.NewInMemoryStorage()
		campaignClient     = NewCampaignClient(nil, nil, nil, mockStorage, mockCampaignHelper)

		businessSummary = researcher.BusinessSummary{BusinessName: "business1", ID: "user1"}
		userID          = "user1"
		mockPages       = []researcher.PageContents{{
			Url: "url1",
		}}

		campaignThemes = []campaign_helper.CampaignTheme{
			{
				PrimaryKeyword: "keyword1",
			},
			{
				PrimaryKeyword: "keyword2",
			},
		}
	)
	mockCampaignHelper.GetCandidatePageContentsForUserWillReturn(userID, mockPages)
	mockCampaignHelper.GenerateThemesWillReturn(businessSummary.BusinessName, campaignThemes)

	// when
	storage.Store(mockStorage, businessSummary)
	themes, err := campaignClient.GenerateThemesForUser(userID)

	// then
	assert.NoError(t, err)
	assert.Equal(t, campaignThemes, themes)
}
