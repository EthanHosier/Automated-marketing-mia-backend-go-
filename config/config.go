package config

import (
	"os"

	"github.com/ethanhosier/mia-backend-go/campaigns"
	"github.com/ethanhosier/mia-backend-go/campaigns/campaign_helper"
	"github.com/ethanhosier/mia-backend-go/canva"
	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/images"
	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/researcher"
	"github.com/ethanhosier/mia-backend-go/services"
	"github.com/ethanhosier/mia-backend-go/storage"
	supa "github.com/nedpals/supabase-go"
)

type ServerConfig struct {
	Researcher     researcher.Researcher
	CampaignClient *campaigns.CampaignClient
	Store          storage.Storage
	ImagesClient   images.ImagesClient
}

func NewProdServerConfig() ServerConfig {
	var (
		httpClient     = &http.HttpClient{}
		canvaClient    = canva.NewClient(os.Getenv("CANVA_CLIENT_ID"), os.Getenv("CANVA_CLIENT_SECRET"), "./canva/canva-tokens.json", httpClient, 300)
		openaiClient   = openai.NewOpenaiClient(os.Getenv("OPENAI_KEY"))
		servicesClient = services.NewServicesClient(httpClient)
		storageClient  = storage.NewSupabaseStorage(newSupabaseClient(), os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SERVICE_KEY"), httpClient)

		r               = researcher.New(servicesClient, openaiClient)
		campaign_helper = campaign_helper.NewCampaignHelperClient(openaiClient, r, canvaClient, storageClient)
		c               = campaigns.NewCampaignClient(openaiClient, r, canvaClient, storageClient, campaign_helper)
		imagesClient    = images.NewHttpImageClient(httpClient, storageClient, openaiClient)
	)

	return ServerConfig{
		Researcher:     r,
		CampaignClient: c,
		Store:          storageClient,
		ImagesClient:   imagesClient,
	}
}

func newSupabaseClient() *supa.Client {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	return supa.CreateClient(supabaseUrl, supabaseServiceKey)
}
