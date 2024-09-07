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
)
