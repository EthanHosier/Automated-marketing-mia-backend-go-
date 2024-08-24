package storage

import "github.com/ethanhosier/mia-backend-go/types"

type Storage interface {
	StoreBusinessSummary(userId string, businessSummary types.BusinessSummary) error
	GetBusinessSummary(userId string) (types.StoredBusinessSummary, error)
	StoreSitemap(userId string, urls []string, embeddings []types.Vector) error
	GetSitemap(userId string) ([]types.StoredSitemapUrl, error)
	GetNearestTemplate(types.Vector) (string, error)
}
