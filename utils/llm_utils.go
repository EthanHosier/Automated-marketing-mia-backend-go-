package utils

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/ethanhosier/mia-backend-go/types"
)

func RunBedrock(prompt string) (string, error) {
	region := os.Getenv("AWS_REGION")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))

	if err != nil {
		return "", err
	}

	brc := bedrockruntime.NewFromConfig(cfg)
	payload := types.BedrockRequest{
		Prompt:      prompt,
		MaxTokens:   512,
		Temperature: 0.5,
	}

	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	output, err := brc.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
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
