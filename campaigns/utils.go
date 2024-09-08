package campaigns

import (
	"context"
	"encoding/json"

	"github.com/ethanhosier/mia-backend-go/openai"
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
	err = json.Unmarshal([]byte(extractedArr), themes)
	if err != nil {
		return nil, err
	}

	return themes, nil
}
