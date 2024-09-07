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

func (s *SupabaseStorage) Store(table TableName, data interface{}) (interface{}, error) {
	tableName, ok := tableNames[table]
	if !ok {
		return nil, errors.New("table not found")
	}

	var results []interface{}
	err := s.client.DB.From(tableName).Insert(data).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) Get(table TableName, id string) (interface{}, error) {
	tableName, ok := tableNames[table]
	if !ok {
		return nil, errors.New("table not found")
	}

	var result interface{}
	err := s.client.DB.From(tableName).Select("*").Single().Eq("id", id).Execute(&result)

	return result, err
}

func (s *SupabaseStorage) GetRandom(table TableName, limit int) ([]interface{}, error) {
	tableName, ok := tableNames[table]
	if !ok {
		return nil, errors.New("table not found")
	}

	var results []interface{}
	err := s.client.DB.From(tableName).Select("*").Limit(limit).Execute(&results)

	rand.Shuffle(len(results), func(i, j int) {
		results[i], results[j] = results[j], results[i]
	})

	l := min(len(results), limit)
	return results[:l], err
}

func (s *SupabaseStorage) Update(table TableName, id string, updateFields map[string]interface{}) (interface{}, error) {
	tableName, ok := tableNames[table]
	if !ok {
		return nil, errors.New("table not found")
	}

	var results []interface{}
	err := s.client.DB.From(tableName).Update(updateFields).Eq("id", id).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) Rpc(rpcMethod RpcMethod, payload map[string]interface{}) (interface{}, error) {
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
