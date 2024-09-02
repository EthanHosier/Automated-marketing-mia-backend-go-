package storage

import "github.com/ethanhosier/mia-backend-go/types"

type Storage interface {
	StoreBusinessSummary(userId string, businessSummary types.BusinessSummary) error
	GetBusinessSummary(userId string) (types.StoredBusinessSummary, error)
	UpdateBusinessSummary(userId string, updateFields map[string]interface{}) error
	StoreSitemap(userId string, urls []string, embeddings []types.Vector) error
	GetSitemap(userId string) ([]types.StoredSitemapUrl, error)
	GetNearestTemplate(types.Vector) (*types.NearestTemplateResponse, error)
	GetNearestUrl(types.Vector) (string, error)
	GetRandomUrls(string, int) ([]string, error)
	GetRandomTemplates(int) ([]types.NearestTemplateResponse, error)
}
