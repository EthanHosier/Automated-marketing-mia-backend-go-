package storage

import "github.com/ethanhosier/mia-backend-go/types"

type TableName int
type RpcMethod int

const (
	CANVA_TEMPLATES TableName = iota
	BUSINESS_SUMMARIES
	SITEMAPS

	NEAREST_TEMPLATE RpcMethod = iota
	NEAREST_URL
	RANDOM_URLS
)

var tableNames = map[TableName]string{
	CANVA_TEMPLATES:    "canva_templates",
	BUSINESS_SUMMARIES: "businessSummaries",
	SITEMAPS:           "sitemaps",
}

var rpcMethods = map[RpcMethod]string{
	RANDOM_URLS:      "/rest/v1/rpc/random_urls",
	NEAREST_URL:      "/rest/v1/rpc/nearest_url",
	NEAREST_TEMPLATE: "/rest/v1/rpc/match_canva_templates",
}

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

	Store(table TableName, data interface{}) (interface{}, error)
	Get(table TableName, id string) (interface{}, error)
	GetRandom(table TableName, limit int) ([]interface{}, error)
	Update(table TableName, id string, updateFields map[string]interface{}) (interface{}, error)

	Rpc(method RpcMethod, payload map[string]interface{}) (interface{}, error)
}
