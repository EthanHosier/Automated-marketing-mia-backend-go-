package images

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/storage"
	"github.com/ethanhosier/mia-backend-go/utils"
)

type ImagesClient interface {
	CaptionsFor(image string) ([]string, error)
	AiImageFrom(prompt string, model AiImageModel) ([]byte, error)
	StockImageFrom(prompt string) (string, error)
	BestImageFor(ctxt context.Context, desiredFeatures []string, guaranteedImages []string, relevanceDescription string, prompt string) (string, error)
	FilterTooSmallImages(images []string) ([]string, error)
}

type HttpImageClient struct {
	httpClient   http.Client
	store        storage.Storage
	openaiClient openai.OpenaiClient
}

func NewHttpImageClient(httpClient http.Client, store storage.Storage, openaiClient openai.OpenaiClient) *HttpImageClient {
	return &HttpImageClient{
		httpClient:   httpClient,
		store:        store,
		openaiClient: openaiClient,
	}
}

func (ic *HttpImageClient) CaptionsFor(image string) ([]string, error) {
	return utils.Retry(3, func() ([]string, error) {
		return ic.getCaptionsCompletionArr(image)
	})
}

// TODO: this is gonna obliterate the rate limit
func (ic *HttpImageClient) CaptionsForAll(images []string) ([][]string, error) {
	tasks := utils.DoAsyncList(images, func(image string) ([]string, error) {
		return ic.getCaptionsCompletionArr(image)
	})

	return utils.GetAsyncList(tasks)
}

func (ic *HttpImageClient) AiImageFrom(prompt string, model AiImageModel) ([]byte, error) {
	var (
		url          = "https://api.stability.ai/v2beta/stable-image/generate/" + string(model)
		outputFormat = "png"

		body   = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
	)

	_ = writer.WriteField("prompt", prompt)
	_ = writer.WriteField("output_format", outputFormat)

	err := writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing writer: %v", err)
	}

	req, err := ic.httpClient.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("STABILITY_API_KEY")))
	req.Header.Set("Accept", "image/*")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ic.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	return respBody, nil
}

func (ic *HttpImageClient) StockImageFrom(prompt string) (string, error) {
	return "", nil
}

// TODO: add some max for number of images tested?
func (ic *HttpImageClient) BestImageFor(ctxt context.Context, desiredFeatures []string, guaranteedImages []string, relevanceDescription string, prompt string) (string, error) {
	if len(guaranteedImages) > 25 {
		slog.Warn("Too many guaranteed images, truncating to 25")
		guaranteedImages = guaranteedImages[:25]
	}

	embeddings, err := ic.openaiClient.Embeddings(desiredFeatures)
	if err != nil {
		return "", err
	}

	tasks := utils.DoAsyncList(embeddings, func(embedding []float32) ([]storage.Similarity[storage.ImageFeature], error) {
		return storage.GetClosest[storage.ImageFeature](ctxt, ic.store, embedding, 3)
	})

	results, err := utils.GetAsyncList(tasks)
	if err != nil {
		return "", err
	}

	allImages := guaranteedImages
	for _, f := range utils.Flatten(results) {
		allImages = append(allImages, f.Item.ImageUrl)
	}

	uniqueImages := utils.RemoveDuplicates(allImages)

	if len(allImages) == 0 {
		return ic.base64AiImageFrom(prompt)
	}

	imgPrompt := fmt.Sprintf(bestImagePrompt, relevanceDescription)

	index, err := utils.Retry(3, func() (int, error) {
		i, err := ic.openaiClient.ImageCompletion(ctxt, imgPrompt, uniqueImages, openai.GPT4o)
		if err != nil {
			return 0, err
		}
		return utils.FirstNumberInString(i)
	})

	if err != nil {
		return "", err
	}

	if index == -1 {
		return ic.base64AiImageFrom(prompt)
	}

	return uniqueImages[index], nil
}

func (ic *HttpImageClient) FilterTooSmallImages(images []string) ([]string, error) {
	var (
		filteredImages []string
	)

	for _, img := range images {
		isSmall, err := ic.isImageBelow400FromURL(img)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error checking image size: %v", err))
			continue
		}

		if !isSmall {
			filteredImages = append(filteredImages, img)
		}
	}

	return filteredImages, nil
}

func (ic *HttpImageClient) base64AiImageFrom(prompt string) (string, error) {
	img, err := ic.AiImageFrom(prompt, openai.GPT4o)
	if err != nil {
		return "", err
	}

	return EncodeToBase64WithMIME(img, "image/png"), nil
}
