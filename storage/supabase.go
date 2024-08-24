package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

func (s *SupabaseStorage) StoreSitemap(userId string, urls []string, embeddings []types.Vector) error {
	uniqueUrls := utils.RemoveDuplicates(urls)

	var rows []types.StoredSitemapUrl
	for i, url := range uniqueUrls {
		rows = append(rows, types.StoredSitemapUrl{
			ID:     userId,
			Url:    url,
			Vector: embeddings[i],
		})
	}

	var results []types.StoredBusinessSummary

	err := s.client.DB.From("sitemaps").Insert(rows).Execute(&results)

	return err
}

func (s *SupabaseStorage) GetBusinessSummary(userId string) (types.StoredBusinessSummary, error) {
	var result types.StoredBusinessSummary
	err := s.client.DB.From("businessSummaries").Select("*").Single().Eq("id", userId).Execute(&result)

	return result, err
}

func (s *SupabaseStorage) GetSitemap(userId string) ([]types.StoredSitemapUrl, error) {
	var results []types.StoredSitemapUrl
	err := s.client.DB.From("sitemaps").Select("*").Eq("id", userId).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) GetNearestTemplate(vector types.Vector) (string, error) {
	url := os.Getenv("SUPABASE_URL") + "/rest/v1/rpc/match_canva_templates"

	// Create a map for the JSON payload
	payload := map[string]interface{}{
		"query_embedding": vector,
		"match_threshold": 0.0,
		"match_count":     10,
	}

	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		return "", err
	}

	fmt.Println(payloadBytes)

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", os.Getenv("SUPABASE_SERVICE_KEY"))

	// Create a new HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return "", err
	}
	defer resp.Body.Close() // Ensure the response body is closed

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	// Print the raw response body
	fmt.Println("Raw response body:", string(body))

	// Parse the JSON response if needed
	var result map[string]interface{} // Adjust based on your response structure
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return "", err
	}

	return "", nil
}
