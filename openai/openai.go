package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sashabaranov/go-openai"
)

type OpenaiClient struct {
	openaiClient *openai.Client
	usageCh      chan openai.Usage
}

func NewOpenaiClient(apiKey string) *OpenaiClient {
	var (
		openaiClient = openai.NewClient(apiKey)
		usageCh      = make(chan openai.Usage)
	)

	go UsageLoop(usageCh)

	return &OpenaiClient{
		openaiClient: openaiClient,
		usageCh:      usageCh,
	}
}

func UsageLoop(usageCh chan openai.Usage) {
	for usage := range usageCh {
		slog.Info("Openai Usage", "prompt_tokens", usage.PromptTokens, "completion_tokens", usage.CompletionTokens, "total_tokens", usage.TotalTokens)
	}
}

func (oc *OpenaiClient) ChatCompletion(ctx context.Context, prompt string, model string) (string, error) {
	resp, err := oc.openaiClient.CreateChatCompletion(
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

	oc.usageCh <- resp.Usage
	return resp.Choices[0].Message.Content, nil
}

func (oc *OpenaiClient) ImageCompletion(ctx context.Context, prompt string, images []string, model string) (string, error) {
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

	resp, err := oc.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)

	if err != nil {
		return "", err
	}

	oc.usageCh <- resp.Usage
	return resp.Choices[0].Message.Content, nil
}

func (oc *OpenaiClient) Embeddings(urls []string) ([][]float32, error) {
	queryReq := openai.EmbeddingRequest{
		Input: urls,
		Model: openai.SmallEmbedding3,
	}

	queryResponse, err := oc.openaiClient.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return nil, fmt.Errorf("error creating query embedding: %w", err)
	}

	queryEmbedding := queryResponse.Data

	var embeddings [][]float32
	for _, embedding := range queryEmbedding {
		embeddings = append(embeddings, embedding.Embedding)
	}

	oc.usageCh <- queryResponse.Usage
	return embeddings, nil
}
