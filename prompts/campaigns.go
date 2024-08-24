package prompts

import (
	"fmt"

	"github.com/ethanhosier/mia-backend-go/types"
)

const (
	themeGeneration = `
	You are a Marketing Director Skilled at deciding high level content marketing themes and keywords for a business.

**Task**: Create 5 brand new marketing themes for the client for this week

**Expected output**:A list of 5 JSON objects with each JSON object containg details of one theme. The JSON object should have the below format:

[{
"theme": string // in under 7 words,
"keywords": string[] // 20 keywords for the theme, ensure these are a mix of small and long keywords. They should have sufficient search volume with low competition for the target location of the business and be SEO friendly,
urlDescription: string // a short description of the url which should be picked to be linked back to the theme
linkedInPostDescription: string // a concise description of the visual elements of the LinkedIn post image. Include details such as color scheme, layout, type of imagery (e.g., photo, illustration, icon), and any specific design features. The description should tie back to the theme and be specific enough to facilitate a vector search match with Canva templates.
instagramPostDescription: string // same as above but for Instagram
twitterXPostDescription: string // same as above but for Twitter
facebookPostDescription: string // same as above but for Facebook
whatsAppPostDescription: string // same as above but for WhatsApp
}]

Respond with just the JSON objects, and no text before or after the opening and closing square brackets.
 

Here are the client details you are currently working for:
%+v

The region they are targeting for these campaigns is: %s


Use these additional instructions to generate the theme, with high priority:
%v

Here are descriptions of images which the user has provided for the theme generation:
%v
`
)

func ThemePrompt(businessSummary types.StoredBusinessSummary, targetAudience string, additionalInstructions string, backlink string, imageDescriptions []string) string {
	ta := targetAudiencePrompt(targetAudience, businessSummary.TargetAudience)

	return fmt.Sprintf(themeGeneration, businessSummary, ta, additionalInstructions, imageDescriptions)
}

func targetAudiencePrompt(ta1 string, ta2 string) string {
	if ta1 == "" {
		return ta1
	}

	return ta2
}
