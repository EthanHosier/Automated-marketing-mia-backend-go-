package campaigns

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/ethanhosier/mia-backend-go/campaigns/campaign_helper"
// 	"github.com/ethanhosier/mia-backend-go/canva"
// 	"github.com/ethanhosier/mia-backend-go/openai"
// 	"github.com/ethanhosier/mia-backend-go/researcher"
// 	"github.com/ethanhosier/mia-backend-go/services"
// 	"github.com/ethanhosier/mia-backend-go/storage"
// 	"github.com/stretchr/testify/assert"
// )

// func TestGenerateThemesForUser(t *testing.T) {
// 	// given
// 	var (
// 		mockCampaignHelper = campaign_helper.NewMockCampaignHelper()
// 		mockStorage        = storage.NewInMemoryStorage()
// 		campaignClient     = NewCampaignClient(nil, nil, nil, mockStorage, mockCampaignHelper)

// 		businessSummary = researcher.BusinessSummary{BusinessName: "business1", ID: "user1"}
// 		userID          = "user1"
// 		mockPages       = []researcher.PageContents{{
// 			Url: "url1",
// 		}}

// 		campaignThemes = []campaign_helper.CampaignTheme{
// 			{
// 				PrimaryKeyword: "keyword1",
// 			},
// 			{
// 				PrimaryKeyword: "keyword2",
// 			},
// 		}
// 	)
// 	mockCampaignHelper.GetCandidatePageContentsForUserWillReturn(userID, mockPages)
// 	mockCampaignHelper.GenerateThemesWillReturn(businessSummary.BusinessName, campaignThemes)

// 	// when
// 	storage.Store(mockStorage, businessSummary)
// 	themes, err := campaignClient.GenerateThemesForUser(userID)

// 	// then
// 	assert.NoError(t, err)
// 	assert.Equal(t, campaignThemes, themes)
// }

// func TestCampaignFrom(t *testing.T) {
// 	var (
// 		mockCampaignHelper = campaign_helper.NewMockCampaignHelper()
// 		mockStorage        = storage.NewInMemoryStorage()
// 		openaiClient       = openai.MockOpenaiClient{}
// 		mockResearcher     = researcher.NewMockResearcher()
// 		canvaClient        = canva.MockCanvaClient{}

// 		campaignClient  = NewCampaignClient(&openaiClient, mockResearcher, &canvaClient, mockStorage, mockCampaignHelper)
// 		userID          = "user1"
// 		businessSummary = researcher.BusinessSummary{BusinessName: "business1", ID: userID}

// 		theme = campaign_helper.CampaignTheme{
// 			PrimaryKeyword:   "keyword1",
// 			SecondaryKeyword: "keyword2",
// 			Url:              "url1",
// 		}

// 		pageBodyText = "page body text"
// 		pageContents = researcher.PageContents{
// 			TextContents: services.WebsiteData{
// 				Title: "title1",
// 			},
// 			Url:       "url1",
// 			ImageUrls: []string{"image1"},
// 		}

// 		socialMediaPosts = []researcher.SocialMediaPost{
// 			{
// 				Platform: researcher.Instagram,
// 				Content:  "content1",
// 				Keyword:  "keyword",
// 			},
// 			{
// 				Platform: researcher.Facebook,
// 				Content:  "content2",
// 				Keyword:  "keyword",
// 			},
// 			{
// 				Platform: researcher.LinkedIn,
// 				Content:  "content3",
// 				Keyword:  "keyword",
// 			},
// 			{
// 				Platform: researcher.Google,
// 				Content:  "content4",
// 				Keyword:  "keyword",
// 			},
// 			{
// 				Platform: researcher.News,
// 				Content:  "content5",
// 				Keyword:  "keyword",
// 			},
// 		}

// 		researchReport = "research report"

// 		templates = []storage.Template{
// 			{
// 				ID: "template1",
// 			},
// 			{
// 				ID: "template2",
// 			},
// 			{
// 				ID: "template3",
// 			},
// 			{
// 				ID: "template4",
// 			},
// 			{
// 				ID: "template5",
// 			},
// 		}

// 		campaignDetailsStr = fmt.Sprintf("Primary keyword: %v\nSecondary keyword: %v\nURL: %v\nTheme: %v\nTemplate Description: %v", theme.PrimaryKeyword, theme.SecondaryKeyword, theme.Url, theme.Theme, theme.ImageCanvaTemplateDescription)

// 		updateTemplateResult = canva.UpdateTemplateResult{
// 			Design: canva.Design{
// 				ID: "design1",
// 			},
// 		}

// 		extractedTemplate = campaign_helper.ExtractedTemplate{
// 			Caption: "caption1",
// 		}
// 	)

// 	mockResearcher.PageBodyTextForWillReturn(theme.Theme, pageBodyText, nil)
// 	mockResearcher.PageContentsForWillReturn(theme.Theme, &pageContents, nil)
// 	mockResearcher.SocialMediaPostsForWillReturn(theme.Theme, socialMediaPosts, nil)
// 	mockResearcher.ResearchReportFromPostsWillReturn(socialMediaPosts, researchReport, nil)

// 	// when
// 	storage.StoreAll(mockStorage, templates)
// 	canvaDesigns,  err := campaignClient.CampaignFrom(theme, &businessSummary)
// }
