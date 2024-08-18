package prompts

const (
	ScrapedWebPageSummary = "Summarise what we learn about the business from this: "

	BusinessSummary = `You are a business analyst Skilled at summarizing large amounts of information into concise paragraphs.

**Task**: Analyze the web scraped pages data and provide the following insights in a structured format:
          1. The business name.
          2. A concise Business Summary in under 150 words.
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
