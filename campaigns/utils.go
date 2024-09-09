package campaigns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

type themeWithSuggestedKeywords struct {
	Theme                         string   `json:"theme"`
	Keywords                      []string `json:"keywords"`
	Url                           string   `json:"url"`
	SelectedUrl                   string   `json:"selectedUrl"`
	ImageCanvaTemplateDescription string   `json:"imageCanvaTemplateDescription"`
}

func (c *CampaignClient) themes(themePrompt string) ([]themeWithSuggestedKeywords, error) {
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

func templatePrompt(platform researcher.SocialMediaPlatform, businessSummary researcher.BusinessSummary, theme string, primaryKeyword string, secondaryKeyword string, url string, scrapedPageBodyText string, scrapedSocialMediaPosts []researcher.SocialMediaPost, fields []storage.TemplateFields, colorFields []storage.ColorField) string {

	relevantSocialMediaPosts := []researcher.SocialMediaPost{}
	for _, smp := range scrapedSocialMediaPosts {
		if smp.Platform == platform || smp.Platform == "news" || smp.Platform == "google" {
			relevantSocialMediaPosts = append(relevantSocialMediaPosts, smp)
		}
	}

	spbt := utils.FirstNChars(scrapedPageBodyText, maxScrapedPageBodyTextCharCount)

	return fmt.Sprintf(openai.PopulateTemplatePlanPrompt, platform, businessSummary, theme, primaryKeyword, secondaryKeyword, platform, primaryKeyword, url, spbt, primaryKeyword, relevantSocialMediaPosts, fields, colorFields, businessSummary.Colors)
}
