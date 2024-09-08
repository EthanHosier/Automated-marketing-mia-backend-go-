package campaigns

import (
	"context"
	"encoding/json"

	"github.com/ethanhosier/mia-backend-go/openai"
)

func (c *CampaignClient) themes(themePrompt string) ([]CampaignTheme, error) {
	completion, err := c.openaiClient.ChatCompletion(context.TODO(), themePrompt, openai.GPT4oMini)

	if err != nil {
		return nil, err
	}

	extractedArr := openai.ExtractJsonData(completion, openai.JSONArray)

	var themes []CampaignTheme
	err = json.Unmarshal([]byte(extractedArr), themes)
	if err != nil {
		return nil, err
	}

	return themes, nil
}
