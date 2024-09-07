package business

import (
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/services"
)

type BusinessClient struct {
	openaiClient   *openai.OpenaiClient
	servicesClient *services.ServicesClient
}

func New(oc *openai.OpenaiClient, sc *services.ServicesClient) *BusinessClient {
	return &BusinessClient{
		openaiClient:   oc,
		servicesClient: sc,
	}
}
