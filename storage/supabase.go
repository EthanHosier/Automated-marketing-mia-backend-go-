package storage

import (
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/ethanhosier/mia-backend-go/utils"
	supa "github.com/nedpals/supabase-go"
)

type SupabaseStorage struct {
	client *supa.Client
}

func NewSupabaseStorage(client *supa.Client) *SupabaseStorage {
	return &SupabaseStorage{
		client: client,
	}
}

func (s *SupabaseStorage) StoreBusinessSummary(userId string, businessSummary types.BusinessSummary) error {
	row := types.StoredBusinessSummary{
		ID:              userId,
		BusinessSummary: businessSummary.BusinessSummary,
		BrandVoice:      businessSummary.BrandVoice,
		TargetRegion:    businessSummary.TargetRegion,
		TargetAudience:  businessSummary.TargetAudience,
	}

	var results []types.StoredBusinessSummary
	err := s.client.DB.From("businessSummaries").Insert(row).Execute(&results)

	return err
}

func (s *SupabaseStorage) StoreSitemap(userId string, urls []string) error {
	uniqueUrls := utils.RemoveDuplicates(urls)

	var rows []types.StoredSitemapUrl
	for _, url := range uniqueUrls {
		rows = append(rows, types.StoredSitemapUrl{
			ID:  userId,
			Url: url,
		})
	}

	var results []types.StoredBusinessSummary

	err := s.client.DB.From("sitemapUrls").Insert(rows).Execute(&results)

	return err
}
