package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	postgrest_go "github.com/nedpals/postgrest-go/pkg"
	supa "github.com/nedpals/supabase-go"
)

type SupabaseStorage struct {
	client        *supa.Client
	url           string
	serviceKey    string
	rpcHttpClient *http.Client
}

func NewSupabaseStorage(client *supa.Client, url string, serviceKey string, rpcHttpClient *http.Client) *SupabaseStorage {

	return &SupabaseStorage{
		client:        client,
		url:           url,
		serviceKey:    serviceKey,
		rpcHttpClient: rpcHttpClient,
	}
}

func (s *SupabaseStorage) store(table string, data interface{}) (interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(table).Insert(data).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) get(table string, id string) (interface{}, error) {
	var result interface{}
	err := s.client.DB.From(table).Select("*").Single().Eq("id", id).Execute(&result)

	return result, err
}

func (s *SupabaseStorage) getRandom(table string, limit int) ([]interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(table).Select("*").Limit(limit).Execute(&results)

	rand.Shuffle(len(results), func(i, j int) {
		results[i], results[j] = results[j], results[i]
	})

	l := min(len(results), limit)
	return results[:l], err
}

func (s *SupabaseStorage) getAll(table string, matchingFields map[string]string) ([]interface{}, error) {
	var results []interface{}

	initial_query := s.client.DB.From(table).Select("*")
	var query *postgrest_go.FilterRequestBuilder
	for k, v := range matchingFields {
		query = initial_query.Eq(k, v)
	}

	err := query.Execute(&results)

	return results, err
}

func (s *SupabaseStorage) update(table string, id string, updateFields map[string]interface{}) (interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(table).Update(updateFields).Eq("id", id).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) rpc(rpcMethod RpcMethod, payload map[string]interface{}) (interface{}, error) {
	method, ok := rpcMethods[rpcMethod]
	if !ok {
		return nil, errors.New("rpc method not found")
	}

	url := s.url + method

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshalling payload:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
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
