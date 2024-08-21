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

func (llm *LLMClient) LlamaSummarise(prompt string) (string, error) {

	payload := types.BedrockRequest{
		Prompt:      prompt,
		MaxTokens:   200,
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

func (llm *LLMClient) OpenaiCompletion(prompt string) (string, error) {
	resp, err := llm.OpenaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o20240806,
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

func (llm *LLMClient) OpenaiEmbeddings(text string) ([]float32, error) {
	queryReq := openai.EmbeddingRequest{
		Input: []string{"text"},
		Model: openai.SmallEmbedding3,
	}

	queryResponse, err := llm.OpenaiClient.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return []float32{}, fmt.Errorf("error creating query embedding: %w", err)
	}

	queryEmbedding := queryResponse.Data[0]

	return queryEmbedding.Embedding, nil
}
