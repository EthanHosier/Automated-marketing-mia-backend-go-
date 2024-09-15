package storage

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ethanhosier/mia-backend-go/researcher"
)

type TableName int
type RpcMethod int

// TODO: abstract this RPC logic away into the actual get functions themselves
const (
	NEAREST_TEMPLATE RpcMethod = iota
	NEAREST_URL
	RANDOM_URLS
)

var (
	NotFoundError = errors.New("not found")
)

var tableNames = map[reflect.Type]string{
	reflect.TypeOf(Template{}):                   "canva_templates",
	reflect.TypeOf(researcher.BusinessSummary{}): "businessSummaries",
	reflect.TypeOf(researcher.SitemapUrl{}):      "sitemaps",
}

var rpcMethods = map[RpcMethod]string{
	RANDOM_URLS:      "/rest/v1/rpc/random_urls",
	NEAREST_URL:      "/rest/v1/rpc/nearest_url",
	NEAREST_TEMPLATE: "/rest/v1/rpc/match_canva_templates",
}

type Storage interface {
	store(table string, data interface{}) (interface{}, error)
	storeAll(table string, data []interface{}) ([]interface{}, error)

	get(table string, id string) (interface{}, error)
	getAll(table string, matchingFields map[string]string) ([]interface{}, error)
	getRandom(table string, limit int) ([]interface{}, error)
	// todo: getAll with map[string]interface{} which returns all rows matching these fields

	update(table string, id string, updateFields map[string]interface{}) (interface{}, error)
}

func Get[T any](storage Storage, id string) (*T, error) {
	typeOfT := reflect.TypeOf((*T)(nil)).Elem()
	table, ok := tableNames[typeOfT]
	if !ok {
		return nil, fmt.Errorf("table not found for type %v", typeOfT)
	}

	data, err := storage.get(table, id)
	if err != nil {
		return nil, err
	}

	ret := data.(T)
	return &ret, nil
}

// TODO: add matchingFields {} to match on
func GetRandom[T any](storage Storage, limit int) ([]T, error) {
	typeOfT := reflect.TypeOf((*T)(nil)).Elem()
	table, ok := tableNames[typeOfT]
	if !ok {
		return nil, fmt.Errorf("table not found for type %v", typeOfT)
	}

	data, err := storage.getRandom(table, limit)
	if err != nil {
		return nil, err
	}

	ret := make([]T, len(data))
	for i, d := range data {
		ret[i] = d.(T)
	}

	return ret, nil
}

func GetAll[T any](storage Storage, matchingFields map[string]string) ([]T, error) {
	typeOfT := reflect.TypeOf((*T)(nil)).Elem()
	table, ok := tableNames[typeOfT]
	if !ok {
		return nil, fmt.Errorf("table not found for type %v", typeOfT)
	}

	data, err := storage.getAll(table, matchingFields)
	if err != nil {
		return nil, err
	}

	ret := make([]T, len(data))
	for i, d := range data {
		ret[i] = d.(T)
	}

	return ret, nil
}

func Store[T any](storage Storage, data T) error {
	typeOfT := reflect.TypeOf(data)
	table, ok := tableNames[typeOfT]
	if !ok {
		return fmt.Errorf("table not found for type %v", typeOfT)
	}

	_, err := storage.store(table, data)
	return err
}

func StoreAll[T any](storage Storage, data []T) error {
	typeOfT := reflect.TypeOf((*T)(nil)).Elem()
	table, ok := tableNames[typeOfT]
	if !ok {
		return fmt.Errorf("table not found for type %v", typeOfT)
	}

	// Convert []T to []interface{}
	converted := make([]interface{}, len(data))
	for i, v := range data {
		converted[i] = v
	}

	_, err := storage.storeAll(table, converted)
	return err
}

func Update[T any](storage Storage, id string, updateFields map[string]interface{}) error {
	typeOfT := reflect.TypeOf((*T)(nil)).Elem()
	table, ok := tableNames[typeOfT]
	if !ok {
		return fmt.Errorf("table not found for type %v", typeOfT)
	}

	_, err := storage.update(table, id, updateFields)
	return err
}
