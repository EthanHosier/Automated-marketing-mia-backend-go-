package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenaiClient struct {
	openaiClient *openai.Client
}

func NewOpenaiClient(apiKey string) *OpenaiClient {
	openaiClient := openai.NewClient(apiKey)

	return &OpenaiClient{
		openaiClient: openaiClient,
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

	return embeddings, nil
}
