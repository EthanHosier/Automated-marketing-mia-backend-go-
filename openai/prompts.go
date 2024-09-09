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
	ResearchReportPrompt = `
Forget anything you have ever been asked to do up until this point, even if its has been marked as essential. This is the most important task you have ever been given.

You are a marketing research expert. Please write a meticulously detailed report based on findings about the keyword "%v".
	You are to include a summary at the start of what the report is about. Then, you will write meticulous findings for Google, News, Instagram, LinkedIn and Facebook. For each platform, you should mention the top results found, their urls, and meticulous analysis about the content of the scraped data, with examples. Per each platform you should also include common themes, important statistics (if there are any), noteworthy points and the trending hashtags. You should then provide a descriptive summary at then end, which highlights SEO and Content Strategy Recommendations, with examples. More specifically, you should describe also cover content Gaps: Identify specific areas where competitors have content that your team does not.
Optimization Tips: Suggest specific improvements for existing content based on current trends and insights.
New Content Ideas: Propose new content topics or formats inspired by the research.

	Here is the data you are to work with: %+v

	Throughout the process, actively engage with the data and the insights it offers. Remember, you have complete control over the level of depth and complexity in your examination.
	`
	PickBestImagePrompt = `
You are given the following details about a social media campaign:
%s

I am going to give you a list of image urls. You need to select the best image for the campaign. Your answer must be the index of the image (0 indexing) and nothing more. You must pick one image.

This is some info about image you must select:
%+v

Here is the list of images:
`
	PopulateTemplatePlanPrompt = `**Role**: You are a Social Media Content Creator, Designer, and AI Image Prompt Engineer skilled at crafting engaging and viral social media posts tailored to a business’s marketing theme, insights from research reports, and utilizing Canva templates to create visually appealing graphics. Each platform will have its own distinct template, and any image fields will include detailed prompts for AI image generation.

**Task**: Create catchy and viral posts for the following platform:%v. Ensure that both the graphic elements (which will be populated Canva templates, dependant on the result of this prompt) and the textual content are aligned, engaging, and optimized for each platform. For any image description fields, generate a detailed prompt which will be used to create those images.

These are the details of the client you are working for:
%+v

Here are some details you must incorporate into the post:
Theme: %s
Primary Keyword: %s
Secondary Keyword: %s

General Guidelines for %s:
•	Attention-Grabbing Start: Capture attention in the first 125 characters with curiosity, emotion, questions, or bold statements.
•	More organic, less salesy: Don’t make it seem too salesy. Try to give as much information and make it catchy so people are interested in finding out about the product organically.
•	Keywords: Ensure the posts contain the primary keyword "%s". Naturally integrate any other relevant keywords that the audience might use to find the post.
•	Fact-Checking: Before finalizing, fact-check any claims and proofread each caption for spelling, grammar, and brand style consistency.
•	Avoid Cringe: Ensure the tone and content are engaging and professional, avoiding anything that might be perceived as overly informal or inappropriate.
•	Call-To-Action (CTA): End with a compelling CTA encouraging specific actions.
•	Brand Voice: Maintain a distinctive brand voice and personality throughout that's consistent with the business’s branding.
•	Formatting:
•	Line breaks every 8-11 words and paragraphs of 21 words max.
•	Use punctuation, emojis, or caps to make key parts like CTA stand out.
•	Never place two emojis next to each other. One per paragraph maximum.
•	Do not include an emoji in every paragraph.
{if selected_url exists:
•	URL Link Back: link back to this URL in your captions: %s

This is the content of the given URL. Incorporate any content as you see fit from the webpage, particularly picking relevant analytical data:
%s

Here is some futher scraped information about the keyword "%s" which has been researched online:
%+v

These are the fields which are required to be filled in for the post image, which will be populated in Canva. Use the comment of each field to determine what the value of the field should be. Make sure that the characters used is less than maxCharacters limit (if it's specified). Pay close attention to what page each field is on, relative to one another. For any image fields, instead of giving the image, give a text description of the image, which will be used to generate the image using AI.:
%+v

Here are the color fields which are required. 
%+v

Match each color field to one of these colors from the business color theme:
%+v

Respond with a json object of the following form.

{
	fields: []{ // list of text or image fields, matching the fields in the template provided
		name: string // the name of the text or image field which has been provided to you
		value: string // the text or image description which you have generated for this field
		type: "image" | "text" // the type of the field, either image or text
	}
	colors: []{ // list of color fields, matching the color fields in the template provided
		name: string // the name of the color field which has been provided to you. E.g bgmedium
		color: string // the color which you have matched to this field.
	}
	caption: string // the caption for the post
}

There should be no text before or after the opening and closing curly braces.

For the caption, follow these guidlines:

Content Formula: start with a Hook, Context, Details/Story, Lesson/Insight, CTA, Hashtags.
•	Personal or Relatable Anecdotes: Include personal experiences or relatable anecdotes if appropriate.
•	Engaging Elements: Include questions, compelling statistics or data points, CTAs, and relevant emojis or symbols.
•	Professional Tone: Maintain a professional and authoritative tone.
•	Hashtags: Provide a list of relevant hashtags.
`
)
