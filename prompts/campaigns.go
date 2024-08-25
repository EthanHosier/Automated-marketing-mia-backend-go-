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

	researchReport = `You are a marketing research expert.

**Task**: You are tasked with generating a one-page daily marketing research report for a given keyword. You have been given the following data:

for the keyword: "%v"
Top 5 Google Search Results: URLs and brief descriptions.
Top 5 News Articles: Titles, sources, and brief summaries.
Top 5 Instagram Posts: URLs or images, captions, hashtags, and engagement metrics.
Top 5 LinkedIn Posts: URLs or images, post content, and engagement metrics.
Top 5 Facebook Posts: URLs or images, post content, and engagement metrics.

**Description**: Use the following format to structure the report and provide the output:

One-Page Daily Marketing Research Report

1. Executive Summary
Objective: Briefly state the purpose of the research (e.g., "To identify key trends and insights from top online content related to "%v"").
Date: The date of the report.

2. Top Google Search Results
Keyword: "%v"
Top 5 URLs:
URL 1
URL 2
URL 3
URL 4
URL 5
Analysis:
Common Themes: Briefly list the recurring themes or topics.
Important Statistics: Highlight key data points or statistics.
Noteworthy Points: Mention any unique or particularly relevant insights.

3. Top News Articles
Top 5 Articles:
Article Title 1 - Source
Article Title 2 - Source
Article Title 3 - Source
Article Title 4 - Source
Article Title 5 - Source
Analysis:
Major News Trends: Summarize the main trends found across the articles.
Emerging Topics: Identify any new or growing topics of interest.
Significant Quotes: Include any impactful quotes from the articles.

4. Top Social Media Posts

Instagram:
Top 5 Posts: List URLs or embed images.
Analysis:
Trending Hashtags: List the popular hashtags used.
Popular Content Types: Describe the types of content that received high engagement.
Engagement Strategies: Highlight effective strategies observed.

LinkedIn:
Top 5 Posts: List URLs or embed images.
Analysis:
Industry Trends: Summarize the main industry-related insights.
Professional Insights: Highlight key points shared by industry professionals.
High-Performing Formats: Note which types of posts were most effective.

Facebook:
Top 5 Posts: List URLs or embed images.
Analysis:
Community Discussions: Summarize the main topics of discussion.
Popular Topics: Identify which topics received the most engagement.
Effective Engagement Techniques: Highlight successful engagement methods.

5. SEO and Content Strategy Recommendations
Content Gaps: Identify specific areas where competitors have content that your team does not.
Optimization Tips: Suggest specific improvements for existing content based on current trends and insights.
New Content Ideas: Propose new content topics or formats inspired by the research.


**Expected output** : Generate the report using the given data. Ensure the output includes only the report and no other text. Handle any missing or incomplete data gracefully by noting the absence in the appropriate section. Keep the urls as their whole link in the report.


Below are all the search data for the keyword "%+v"`

	researchReport2 = `
You are a marketing research expert. Please write a descriptive report based on findings about the keyword "%v" The report should span ~1000 words.
	You are to include a summary at the start of what the report is about, then desriptive findings for Google, News, Instagram, LinkedIn and Facebook. For each platform, you should list the top results found, their urls, and deap analysis about the content of the scraped data, with examples. Per each platform you should include common themes, important statistics, noteworthy points and Trending Hashtags. You should then provide a descriptive summary at then end, which highlights SEO and Content Strategy Recommendations, with examples. More specifically, you should describe Content Gaps: Identify specific areas where competitors have content that your team does not.
Optimization Tips: Suggest specific improvements for existing content based on current trends and insights.
New Content Ideas: Propose new content topics or formats inspired by the research.

	Here is the data you are to work with: %+v
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

func SummarisePostPrompt(platform string) (string, error) {
	switch platform {
	case "linkedIn":
		return "Summarise the LinkedIn post description: ", nil
	case "instagram":
		return "Summarise the Instagram post description: ", nil
	case "facebook":
		return "Summarise the Facebook post description: ", nil
	case "google":
		return "Summarise the scraped website page content: ", nil
	case "news":
		return "Summarise the news article: ", nil
	default:
		return "", fmt.Errorf("platform %s not supported", platform)
	}
}

func ResearchReportPrompt(keyword string, researchReportData types.ResearchReportData) string {
	return fmt.Sprintf(researchReport, keyword, researchReportData)
}
