package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/ethanhosier/mia-backend-go/http"
	postgrest_go "github.com/nedpals/postgrest-go/pkg"
	supa "github.com/nedpals/supabase-go"
)

var (
	getNearestRpcMethods = map[TableName]string{}
)

type SupabaseStorage struct {
	client        *supa.Client
	url           string
	serviceKey    string
	rpcHttpClient http.Client
}

func NewSupabaseStorage(client *supa.Client, url string, serviceKey string, rpcHttpClient http.Client) *SupabaseStorage {

	return &SupabaseStorage{
		client:        client,
		url:           url,
		serviceKey:    serviceKey,
		rpcHttpClient: rpcHttpClient,
	}
}

func (s *SupabaseStorage) store(table TableName, data interface{}) (interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(string(table)).Insert(data).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) storeAll(table TableName, data []interface{}) ([]interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(string(table)).Insert(data).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) get(table TableName, id string) (interface{}, error) {
	var result []interface{}
	err := s.client.DB.From(string(table)).Select("*").Limit(1).Eq("id", id).Execute(&result)

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, NotFoundError
	}

	return result[0], nil
}

func (s *SupabaseStorage) getRandom(table TableName, limit int) ([]interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(string(table)).Select("*").Execute(&results)

	rand.Shuffle(len(results), func(i, j int) {
		results[i], results[j] = results[j], results[i]
	})

	l := min(len(results), limit)
	return results[:l], err
}

func (s *SupabaseStorage) getAll(table TableName, matchingFields map[string]string) ([]interface{}, error) {
	var results []interface{}

	initial_query := s.client.DB.From(string(table)).Select("*")
	var query *postgrest_go.FilterRequestBuilder
	for k, v := range matchingFields {
		query = initial_query.Eq(k, v)
	}

	err := query.Execute(&results)

	return results, err
}

func (s *SupabaseStorage) update(table TableName, id string, updateFields map[string]interface{}) (interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(string(table)).Update(updateFields).Eq("id", id).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) getClosest(table TableName, vector []uint32, limit int) ([]interface{}, error) {
	return nil, fmt.Errorf("getClosest not implemented for SupabaseStorage")
}

func (s *SupabaseStorage) rpc(rpcMethod string, payload map[string]interface{}) ([]interface{}, error) {
	url := s.url + rpcMethod

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshalling payload:", err)
		return nil, err
	}

	req, err := s.rpcHttpClient.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.serviceKey)

	resp, err := s.rpcHttpClient.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return result, nil
}
