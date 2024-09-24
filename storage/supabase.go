package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"reflect"

	"github.com/ethanhosier/mia-backend-go/http"
	"github.com/ethanhosier/mia-backend-go/utils"
	postgrest_go "github.com/nedpals/postgrest-go/pkg"
	supa "github.com/nedpals/supabase-go"
)

var (
	getNearestRpcMethods = map[TableName]string{
		image_features_table: "/rest/v1/rpc/match_image_features",
	}
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

func (s *SupabaseStorage) getRandom(table TableName, limit int, matchingFields map[string]string) ([]interface{}, error) {
	var results []interface{}
	initialQuery := s.client.DB.From(string(table)).Select("*")

	var err error

	if matchingFields == nil {
		err = initialQuery.Execute(&results)
	} else {
		var query *postgrest_go.FilterRequestBuilder
		for k, v := range matchingFields {
			query = initialQuery.Eq(k, v)
		}
		err = query.Execute(&results)
	}

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
		query = initial_query.Eq(k, v) //todo: this will only work for 1 key value pair
	}

	err := query.Execute(&results)

	return results, err
}

func (s *SupabaseStorage) update(table TableName, id string, updateFields map[string]interface{}) (interface{}, error) {
	var results []interface{}
	err := s.client.DB.From(string(table)).Update(updateFields).Eq("id", id).Execute(&results)

	return results, err
}

func (s *SupabaseStorage) getClosest(ctxt context.Context, table TableName, vector []float32, limit int) ([]Similarity[interface{}], error) {
	userId := ctxt.Value(utils.UserIdKey).(string)
	payload := map[string]interface{}{
		"query_embedding": vector,
		"match_threshold": 0.5,
		"match_count":     limit,
		"user_id":         userId,
	}

	data, err := s.rpc(getNearestRpcMethods[table], payload)
	if err != nil {
		return nil, err
	}

	var results []Similarity[interface{}]
	for _, item := range data {
		similarity, newItem, err := extractAndRemoveSimilarity(item)
		if err != nil {
			return nil, err
		}

		results = append(results, Similarity[interface{}]{Item: newItem, Similarity: similarity})
	}

	return results, nil
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

func extractAndRemoveSimilarity(input interface{}) (float64, interface{}, error) {
	val := reflect.ValueOf(input)
	typ := val.Type()

	// Check if input is a struct
	if val.Kind() != reflect.Struct {
		return 0, input, errors.New("input must be a struct")
	}

	// Find the "Similarity" field
	similarityField := val.FieldByName("Similarity")
	if !similarityField.IsValid() {
		return 0, input, errors.New("field 'Similarity' not found")
	}

	if similarityField.Kind() != reflect.Float64 {
		return 0, input, errors.New("field 'Similarity' is not of type float64")
	}

	// Extract the similarity value
	similarity := similarityField.Float()

	// Create a new struct type without the "Similarity" field
	var fields []reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Name != "Similarity" {
			fields = append(fields, field)
		}
	}

	// Create a new struct value
	newStructType := reflect.StructOf(fields)
	newStructValue := reflect.New(newStructType).Elem()

	// Copy fields to the new struct, excluding "Similarity"
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if typ.Field(i).Name != "Similarity" {
			newStructValue.FieldByName(typ.Field(i).Name).Set(field)
		}
	}

	return similarity, newStructValue.Interface(), nil
}
