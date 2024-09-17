package images

import (
	"bytes"
	"fmt"
	"io"
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
	BestImageFor(desiredFeatures []string, prompt string) (string, error)
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

func (ic *HttpImageClient) BestImageFor(desiredFeatures []string, prompt string) (string, error) {
	_, err := ic.openaiClient.Embeddings(desiredFeatures)
	if err != nil {
		return "", err
	}

	return "", nil
}
