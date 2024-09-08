package openai

const (
	ColorThemesPrompt           = "I am going to give you a screenshot of a business' landing page. Please return me a list of the 5 main hex colors which represent's the business' theme. The response should just be json a list of colours, and nothing else before the start and closed bracket. For examle: [\"#000000\", \"#FFFFFF\"]"
	ScrapedWebPageSummaryPrompt = "Summarise what we learn about the business from this: "

	BusinessSummaryPrompt = `You are a business analyst Skilled at summarizing large amounts of information into concise paragraphs.

**Task**: Analyze the web scraped pages data and provide the following insights in a structured format:
          1. The business name.
          2. A meticulously detailed Business Summary in about 300 words
          3. The brand voice and tone in under 50 words.
          4. Target region of the business.
          5. Target audience / customer profile of the business in under 50 words.
              
**Expected output**: A formatted json object with the below format:
{
    "businessName": "name of business",
	"businessSummary": "business summary in under 150 words",
    "brandVoice": "brand voice in under 50 words",
    "targetRegion": "Target region of the business",
    "targetAudience": "Target customer profile of the business in under 50 words",
 }
 
RESPOND WITH JUST THE JSON OBJECT, and no text before or after the opening and closing curly braces.

Here is the data collected about the business:
`
	ThemeGenerationPrompt = `
	You are a Marketing Director Skilled at deciding high level content marketing themes and keywords for a business.

**Task**: Create 5 brand new marketing themes for the client for this week
 
Here are the client details you are currently working for:
%+v

Here are candidate web pages which have been scraped from the user's website. Each theme must be entirely relevant to ONE of these pages. This includes the url and theme name.
%+v

The region they are targeting for these campaigns is: %s


Use these additional instructions to generate the theme, with high priority:
%v

Here are descriptions of images which the user has provided for the theme generation:
%v


**Expected output**:A list of 5 JSON objects with each JSON object containg details of one theme. The JSON object should have the below format:

[{
"theme": string // in under 7 words,
"keywords": string[] // 20 keywords for the theme, ensure these are a mix of small and long keywords. They should have sufficient search volume with low competition for the target location of the business and be SEO friendly,
url: string // the selected url from the given candidate pages of the user's website which the theme is relevant to,
imageCanvaTemplateDescription: string // a concise description of the visual elements of the post image. Include details such as color scheme, layout, type of imagery (e.g., photo, illustration, icon), and any specific design features. The description should tie back to the theme and be specific enough to facilitate a vector search match with Canva templates.
}]

Respond with just the JSON objects, and no text before or after the opening and closing square brackets.
`
)
