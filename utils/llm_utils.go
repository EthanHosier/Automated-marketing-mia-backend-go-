package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/sashabaranov/go-openai"
)

type LLMClient struct {
	BedrockClient *bedrockruntime.Client
	OpenaiClient  *openai.Client
}

func CreateLLMClient() *LLMClient {
	bedrockClient := createBedrockClient()
	openaiClient := createOpenaiClient()

	return &LLMClient{
		BedrockClient: bedrockClient,
		OpenaiClient:  openaiClient,
	}
}

func createBedrockClient() *bedrockruntime.Client {
	region := os.Getenv("AWS_REGION")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))

	if err != nil {
		log.Fatalf("Failed to create aws bedrock client, error: %v", err)
	}

	return bedrockruntime.NewFromConfig(cfg)
}

func createOpenaiClient() *openai.Client {
	apiKey := os.Getenv("OPENAI_KEY")

	return openai.NewClient(apiKey)
}

func (llm *LLMClient) LlamaSummarise(prompt string, maxTokens int) (string, error) {

	if maxTokens > 1000 {
		return "", fmt.Errorf("max tokens must be less than 1000")
	}

	if maxTokens < 1 {
		return "", fmt.Errorf("max tokens must be greater than 0")
	}

	payload := types.BedrockRequest{
		Prompt:      prompt,
		MaxTokens:   maxTokens,
		Temperature: 0.5,
	}

	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	output, err := llm.BedrockClient.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		Body:        payloadBytes,
		ModelId:     aws.String("meta.llama3-1-8b-instruct-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("*/*"),
	})

	if err != nil {
		return "", err
	}

	var response types.BedrockResponse
	err = json.Unmarshal(output.Body, &response)

	if err != nil {
		return "", err
	}

	return response.Generation, nil
}

func (llm *LLMClient) OpenaiCompletion(prompt string, model string) (string, error) {
	resp, err := llm.OpenaiClient.CreateChatCompletion(
		context.Background(),
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

func (llm *LLMClient) OpenaiEmbeddings(urls []string) ([]types.Vector, error) {
	queryReq := openai.EmbeddingRequest{
		Input: urls,
		Model: openai.SmallEmbedding3,
	}

	queryResponse, err := llm.OpenaiClient.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return []types.Vector{}, fmt.Errorf("error creating query embedding: %w", err)
	}

	queryEmbedding := queryResponse.Data

	var embeddings []types.Vector
	for _, embedding := range queryEmbedding {
		embeddings = append(embeddings, embedding.Embedding)
	}

	return embeddings, nil
}

func (llm *LLMClient) OpenaiEmbedding(text string) (types.Vector, error) {
	queryReq := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	}

	queryResponse, err := llm.OpenaiClient.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return types.Vector{}, fmt.Errorf("error creating query embedding: %w", err)
	}

	return queryResponse.Data[0].Embedding, nil
}
