package images

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ethanhosier/mia-backend-go/openai"
)

func (ic *HttpImageClient) getCaptionsCompletionArr(image string) ([]string, error) {
	result, err := ic.openaiClient.ImageCompletion(context.TODO(), featuresPrompt, []string{image}, openai.GPT4o)
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

func EncodeToBase64WithMIME(data []byte, mimeType string) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}
