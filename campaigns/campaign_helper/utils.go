package campaign_helper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethanhosier/mia-backend-go/openai"
)

func (c *CampaignHelperClient) getCaptionsCompletionArr(imageDescription string) ([]string, error) {
	prompt := fmt.Sprintf(featuresFromDescriptionPrompt, imageDescription)
	result, err := c.openaiClient.ChatCompletion(context.TODO(), prompt, openai.GPT4o)
	if err != nil {
		return nil, err
	}

	extractedJsonString := openai.ExtractJsonData(result, openai.JSONArray)
	var captions []string
	err = json.Unmarshal([]byte(extractedJsonString), &captions)

	if err != nil {
		return nil, err
	}

	return captions, nil
}
