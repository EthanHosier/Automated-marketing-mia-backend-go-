package campaigns

import (
	"fmt"

	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

const (
	maxScrapedPageBodyTextCharCount = 4000
)

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
