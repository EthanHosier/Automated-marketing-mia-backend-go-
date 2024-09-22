package images

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"

	"golang.org/x/image/webp"

	"github.com/ethanhosier/mia-backend-go/openai"
)

func (ic *HttpImageClient) getCaptionsCompletionArr(imageURL string) ([]string, error) {
	result, err := ic.openaiClient.ImageCompletion(context.TODO(), featuresPrompt, []string{imageURL}, openai.GPT4o)
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

func (ic *HttpImageClient) isImageBelow400FromURL(imageURL string) (bool, error) {
	resp, err := ic.httpClient.Get(imageURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check the Content-Type header to determine the image format
	contentType := resp.Header.Get("Content-Type")
	var m image.Image

	switch contentType {
	case "image/webp":
		m, err = webp.Decode(resp.Body)
		if err != nil {
			return false, fmt.Errorf("failed to decode webp image: %v", err)
		}
	default:
		// For other image formats like jpeg, png
		m, _, err = image.Decode(resp.Body)
		if err != nil {
			return false, fmt.Errorf("failed to decode image: %v", err)
		}
	}

	g := m.Bounds()

	// Get height and width
	height := g.Dy()
	width := g.Dx()

	return height < 400 || width < 400, nil
}
