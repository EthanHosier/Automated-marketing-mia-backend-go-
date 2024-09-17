package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/ethanhosier/mia-backend-go/researcher"
)

type TableName string

const (
	canva_templates   TableName = "canva_templates"
	businessSummaries TableName = "businessSummaries"
	sitemaps          TableName = "sitemaps"
	image_features    TableName = "image_features"
)

var (
	NotFoundError = errors.New("not found")
)

var tableNames = map[reflect.Type]TableName{
	reflect.TypeOf(Template{}):                   canva_templates,
	reflect.TypeOf(researcher.BusinessSummary{}): businessSummaries,
	reflect.TypeOf(researcher.SitemapUrl{}):      sitemaps,
	reflect.TypeOf(ImageFeature{}):               image_features,
}

type Storage interface {
	store(table TableName, data interface{}) (interface{}, error)
	storeAll(table TableName, data []interface{}) ([]interface{}, error)

	get(table TableName, id string) (interface{}, error)
	getAll(table TableName, matchingFields map[string]string) ([]interface{}, error)
	getRandom(table TableName, limit int) ([]interface{}, error)
	getClosest(table TableName, vector []uint32, limit int) ([]interface{}, error)
	// todo: getAll with map[string]interface{} which returns all rows matching these fields

	update(table TableName, id string, updateFields map[string]interface{}) (interface{}, error)
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

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to JSON: %v", err)
	}

	ret := new(T)
	err = json.Unmarshal(jsonData, ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data into type %v: %v", typeOfT, err)
	}

	return ret, nil
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
		jsonData, err := json.Marshal(d)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data to JSON: %v", err)
		}

		err = json.Unmarshal(jsonData, &ret[i])
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal data into type %v: %v", typeOfT, err)
		}
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
		jsonData, err := json.Marshal(d)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data to JSON: %v", err)
		}

		err = json.Unmarshal(jsonData, &ret[i])
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal data into type %v: %v", typeOfT, err)
		}
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
