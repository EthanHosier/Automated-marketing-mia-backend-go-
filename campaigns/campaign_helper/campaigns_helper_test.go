package campaign_helper

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethanhosier/mia-backend-go/canva"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestChosenKeywords(t *testing.T) {
	// given
	var (
		r = researcher.NewMockResearcher()

		c = NewCampaignHelperClient(nil, r, nil, nil, nil)

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
		c = NewCampaignHelperClient(nil, r, nil, nil, nil)

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
		c = NewCampaignHelperClient(nil, r, nil, s, nil)

		userID     = "user1"
		sitemapUrl = researcher.SitemapUrl{
			ID:  "id1",
			Url: "url1",
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
	res, err := c.GetCandidatePageContentsForUser(userID, 1)

	// then
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, pageContents[0].Url, res[0].Url)
}

func TestBestImage(t *testing.T) {
	// given
	var (
		op = openai.MockOpenaiClient{}
		c  = NewCampaignHelperClient(&op, nil, nil, nil, nil)

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
		c  = NewCampaignHelperClient(&op, nil, nil, nil, nil)

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

func TestGenerateThemes(t *testing.T) {
	// given
	var (
		op = openai.MockOpenaiClient{}
		r  = researcher.NewMockResearcher()
		c  = NewCampaignHelperClient(&op, r, nil, nil, nil)

		theme1 = themeWithSuggestedKeywords{
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

		adsKeywords1 = []researcher.GoogleAdsKeyword{
			{
				Keyword:            "sleek",
				AvgMonthlySearches: 100,
				CompetitionLevel:   "low",
				CompetitionIndex:   1,
				LowTopOfPageBid:    1,
				HighTopOfPageBid:   2,
			},
		}

		adsKeywords2 = []researcher.GoogleAdsKeyword{
			{
				Keyword:            "retro",
				AvgMonthlySearches: 100,
				CompetitionLevel:   "low",
				CompetitionIndex:   1,
				LowTopOfPageBid:    1,
				HighTopOfPageBid:   2,
			},
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

		pageContents    = []researcher.PageContents{}
		businessSummary = &researcher.BusinessSummary{}
		themePrompt     = fmt.Sprintf(openai.ThemeGenerationPrompt, businessSummary, pageContents, businessSummary.TargetRegion, "", "")
	)

	op.WillReturnChatCompletion(themePrompt, openai.GPT4oMini, themesStr)

	r.GoogleAdsKeywordsDataWillReturn(theme1.Keywords, adsKeywords1, nil)
	r.GoogleAdsKeywordsDataWillReturn(theme2.Keywords, adsKeywords2, nil)

	r.OptimalKeywordsWillReturn(adsKeywords1, "prim1", "sec1", nil)
	r.OptimalKeywordsWillReturn(adsKeywords2, "prim1", "sec2", nil)

	// when
	res, err := c.GenerateThemes(pageContents, businessSummary)

	// then
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, theme1.Theme, res[0].Theme)
	assert.Equal(t, "prim1", res[0].PrimaryKeyword)
	assert.Equal(t, "sec1", res[0].SecondaryKeyword)
	assert.Equal(t, theme1.Url, res[0].Url)
	assert.Equal(t, theme1.SelectedUrl, res[0].SelectedUrl)
	assert.Equal(t, theme1.ImageCanvaTemplateDescription, res[0].ImageCanvaTemplateDescription)

	assert.Equal(t, theme2.Theme, res[1].Theme)
	assert.Equal(t, "prim1", res[1].PrimaryKeyword)
	assert.Equal(t, "sec2", res[1].SecondaryKeyword)
	assert.Equal(t, theme2.Url, res[1].Url)
	assert.Equal(t, theme2.SelectedUrl, res[1].SelectedUrl)
	assert.Equal(t, theme2.ImageCanvaTemplateDescription, res[1].ImageCanvaTemplateDescription)
}

func TestInitColorFields(t *testing.T) {
	// given
	var (
		canvaClient = canva.MockCanvaClient{}
		c           = NewCampaignHelperClient(nil, nil, &canvaClient, nil, nil)

		color1 = "#FFFFFF"
		color2 = "#000000"

		id1 = "id1"
		id2 = "id2"

		colorFields = []PopulatedColorField{
			{
				Name:  "color1",
				Color: color1,
			},
			{
				Name:  "color2",
				Color: color2,
			},
		}
	)
	canvaClient.WillReturnUploadColorAssets([]string{color1, color2}, []string{id1, id2})

	// when
	res, err := c.initColorFields(colorFields)

	// then
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, id1, res[0].ColorAssetId)
	assert.Equal(t, id2, res[1].ColorAssetId)
}

func TestInitImageFields(t *testing.T) {
	// given
	var (
		canvaClient = canva.MockCanvaClient{}
		op          = openai.MockOpenaiClient{}
		c           = NewCampaignHelperClient(&op, nil, &canvaClient, nil, nil)

		imgFields = []PopulatedField{
			{
				Name:  "img1",
				Value: "val1",
				Type:  ImageType,
			},
			{
				Name:  "img2",
				Value: "val2",
				Type:  ImageType,
			},
		}

		candidateImages    = []string{"candidateImg1", "candidateImg2"}
		campaignDetailsStr = "campaignDetails"

		prompt1 = fmt.Sprintf(openai.PickBestImagePrompt, campaignDetailsStr, "val1")
		prompt2 = fmt.Sprintf(openai.PickBestImagePrompt, campaignDetailsStr, "val2")

		imgAssetId1 = "imgAssetId1"
		imgAssetId2 = "imgAssetId2"
	)

	op.WillReturnImageCompletion(prompt1, candidateImages, openai.GPT4o, "0")
	op.WillReturnImageCompletion(prompt2, candidateImages, openai.GPT4o, "1")

	canvaClient.WillReturnUploadImageAssets(candidateImages, []string{imgAssetId1, imgAssetId2})

	// when
	res, err := c.initImageFields(context.TODO(), imgFields, candidateImages, campaignDetailsStr)

	// then
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, imgAssetId1, res[0].AssetId)
	assert.Equal(t, imgAssetId2, res[1].AssetId)
}

func TestInitFields(t *testing.T) {
	// given
	var (
		canvaClient = canva.MockCanvaClient{}
		op          = openai.MockOpenaiClient{}
		c           = NewCampaignHelperClient(&op, nil, &canvaClient, nil, nil)

		imgFields = []PopulatedField{
			{
				Name:  "img1",
				Value: "val1",
				Type:  ImageType,
			},
			{
				Name:  "img2",
				Value: "val2",
				Type:  ImageType,
			},
		}

		colorFields = []PopulatedColorField{
			{
				Name:  "color1",
				Color: "#FFFFFF",
			},
			{
				Name:  "color2",
				Color: "#000000",
			},
		}

		textFields = []PopulatedField{
			{
				Name:  "text1",
				Value: "text val1",
				Type:  TextType,
			},
			{
				Name:  "text2",
				Value: "text val2",
				Type:  TextType,
			},
		}

		extractedTemplate = ExtractedTemplate{
			Platform:    "platform",
			Fields:      append(imgFields, textFields...),
			ColorFields: colorFields,
			Caption:     "caption",
		}

		candidateImages    = []string{"candidateImg1", "candidateImg2"}
		campaignDetailsStr = "campaignDetails"

		prompt1 = fmt.Sprintf(openai.PickBestImagePrompt, campaignDetailsStr, "val1")
		prompt2 = fmt.Sprintf(openai.PickBestImagePrompt, campaignDetailsStr, "val2")

		imgAssetId1 = "imgAssetId1"
		imgAssetId2 = "imgAssetId2"

		color1 = "#FFFFFF"
		color2 = "#000000"

		colorAssetId1 = "colorAssetId1"
		colorAssetId2 = "colorAssetId2"
	)

	op.WillReturnImageCompletion(prompt1, candidateImages, openai.GPT4o, "0")
	op.WillReturnImageCompletion(prompt2, candidateImages, openai.GPT4o, "1")

	canvaClient.WillReturnUploadImageAssets(candidateImages, []string{imgAssetId1, imgAssetId2})
	canvaClient.WillReturnUploadColorAssets([]string{color1, color2}, []string{colorAssetId1, colorAssetId2})

	// when
	textRes, imgRes, colorRes, err := c.InitFields(context.TODO(), &extractedTemplate, campaignDetailsStr, candidateImages)

	// then
	assert.NoError(t, err)
	assert.Len(t, imgRes, 2)
	assert.Equal(t, imgAssetId1, imgRes[0].AssetId)
	assert.Equal(t, imgAssetId2, imgRes[1].AssetId)

	assert.Len(t, colorRes, 2)
	assert.Equal(t, colorAssetId1, colorRes[0].ColorAssetId)
	assert.Equal(t, colorAssetId2, colorRes[1].ColorAssetId)

	assert.Len(t, textRes, 2)
	assert.Equal(t, textFields[0].Value, textRes[0].Text)
	assert.Equal(t, textFields[1].Value, textRes[1].Text)
}

func TestTemplatePlan(t *testing.T) {
	var (
		op = openai.MockOpenaiClient{}
		c  = NewCampaignHelperClient(&op, nil, nil, nil, nil)

		templatePrompt        = "template prompt"
		extractedTemplateJSON = `{
			"platform": "instagram",
			"fields": [
				{
					"name": "field1",
					"value": "value1",
					"type": "text"
				},
				{
					"name": "field2",
					"value": "value2",
					"type": "image"
				}
			],
			"colors": [
				{
					"name": "primaryColor",
					"color": "#FF5733"
				},
				{
					"name": "secondaryColor",
					"color": "#33FF57"
				}
			],
			"caption": "This is a sample caption."
		}`

		extractedTemplate = ExtractedTemplate{
			Platform: "instagram",
			Fields: []PopulatedField{
				{
					Name:  "field1",
					Value: "value1",
					Type:  TextType,
				},
				{
					Name:  "field2",
					Value: "value2",
					Type:  ImageType,
				},
			},
			ColorFields: []PopulatedColorField{
				{
					Name:  "primaryColor",
					Color: "#FF5733",
				},
				{
					Name:  "secondaryColor",
					Color: "#33FF57",
				},
			},
			Caption: "This is a sample caption.",
		}

		templateToFill = storage.Template{
			Fields: []storage.TemplateFields{
				{
					Name:          "field1",
					Type:          "text",
					MaxCharacters: 900,
				},
			},
		}
	)

	// given
	op.WillReturnChatCompletion(templatePrompt, openai.GPT4o, extractedTemplateJSON)

	// when
	res, err := c.TemplatePlan(templatePrompt, templateToFill)

	assert.NoError(t, err)
	assert.Equal(t, *res, extractedTemplate)
}

func TestRephraseTextFieldCharsWithRetry(t *testing.T) {
	// given
	var (
		op = openai.MockOpenaiClient{}
		c  = NewCampaignHelperClient(&op, nil, nil, nil, nil)

		text      = "This is a sample text."
		shortText = "short"
		maxChars  = len(text) - 1

		populatedField = PopulatedField{
			Name:  "field1",
			Value: text,
			Type:  TextType,
		}

		expectedTextFields = PopulatedField{
			Name:  "field1",
			Value: shortText,
			Type:  TextType,
		}
	)

	prompt1 := fmt.Sprintf(openai.MaxCharsPrompt, maxChars, populatedField.Value)

	op.WillReturnChatCompletion(prompt1, openai.GPT4o, shortText)

	// when
	res, err := c.rephraseTextFieldCharsWithRetry(maxChars, &populatedField)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedTextFields, *res)
}
