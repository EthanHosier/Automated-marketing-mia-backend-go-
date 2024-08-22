package utils

import (
	"encoding/json"

	"github.com/ethanhosier/mia-backend-go/types"
)

func Themes(themePrompt string, llmClient *LLMClient) ([]types.ThemeData, error) {
	completion, err := llmClient.OpenaiCompletion(themePrompt)

	if err != nil {
		return nil, err
	}

	extractedArr := ExtractJsonObj(completion, SquareBracket)

	var themeData []types.ThemeData
	err = json.Unmarshal([]byte(extractedArr), &themeData)
	if err != nil {
		return nil, err
	}

	return themeData, nil
}
