package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sashabaranov/go-openai"
)

type OpenaiClient interface {
	ChatCompletion(ctx context.Context, prompt string, model string) (string, error)
	ImageCompletion(ctx context.Context, prompt string, images []string, model string) (string, error)
	Embeddings(urls []string) ([][]float32, error)
}

type GoOpenaiClient struct {
	client  *openai.Client
	usageCh chan UsageData
}

type UsageData struct {
	openai.Usage
	prompt string
}

func NewOpenaiClient(apiKey string) *GoOpenaiClient {
	var (
		openaiClient = openai.NewClient(apiKey)
		usageCh      = make(chan UsageData)
	)

	go usageLoop(usageCh)

	return &GoOpenaiClient{
		client:  openaiClient,
		usageCh: usageCh,
	}
}

func usageLoop(usageCh chan UsageData) {
	for usage := range usageCh {
		slog.Info("Openai request completed", "prompt_tokens", usage.PromptTokens, "completion_tokens", usage.CompletionTokens, "total_tokens", usage.TotalTokens, "prompt", fmt.Sprintf("%s...", usage.prompt[:min(len(usage.prompt), 40)]))
	}
}

func (oc *GoOpenaiClient) ChatCompletion(ctx context.Context, prompt string, model string) (string, error) {
	resp, err := oc.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	oc.usageCh <- UsageData{resp.Usage, prompt}
	return resp.Choices[0].Message.Content, nil
}

func (oc *GoOpenaiClient) ImageCompletion(ctx context.Context, prompt string, images []string, model string) (string, error) {
	imageMessages := []openai.ChatCompletionMessage{}

	for _, image := range images {
		imageMessages = append(imageMessages, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type:     openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{URL: image},
				},
			},
		})
	}

	// Create the full list of messages, starting with the system message
	messages := append([]openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		},
	}, imageMessages...)

	resp, err := oc.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)

	if err != nil {
		return "", err
	}

	oc.usageCh <- UsageData{resp.Usage, prompt}
	return resp.Choices[0].Message.Content, nil
}

func (oc *GoOpenaiClient) Embeddings(urls []string) ([][]float32, error) {
	queryReq := openai.EmbeddingRequest{
		Input: urls,
		Model: openai.SmallEmbedding3,
	}

	queryResponse, err := oc.client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return nil, fmt.Errorf("error creating query embedding: %w", err)
	}

	queryEmbedding := queryResponse.Data

	var embeddings [][]float32
	for _, embedding := range queryEmbedding {
		embeddings = append(embeddings, embedding.Embedding)
	}

	oc.usageCh <- UsageData{queryResponse.Usage, "vector embedding"}
	return embeddings, nil
}
