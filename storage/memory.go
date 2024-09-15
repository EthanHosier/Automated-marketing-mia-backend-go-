package storage

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type InMemoryStorage struct {
	data map[string]map[string]interface{} // Table -> ID -> Data
	mu   sync.RWMutex                      // For concurrent access
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]map[string]interface{}),
	}
}

func (s *InMemoryStorage) store(table string, data interface{}) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use reflection to extract the ID field from data
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	// Find the ID field
	idField := dataValue.FieldByName("ID")
	if !idField.IsValid() || idField.IsZero() {
		return nil, errors.New("ID field not found or is zero")
	}

	id, ok := idField.Interface().(string)
	if !ok {
		return nil, errors.New("ID field is not of type string")
	}

	if _, ok := s.data[table]; !ok {
		s.data[table] = make(map[string]interface{})
	}
	s.data[table][id] = data

	return id, nil
}

func (s *InMemoryStorage) storeAll(table string, data []interface{}) ([]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ids []interface{}
	for _, item := range data {
		// Extract ID from each item
		dataValue := reflect.ValueOf(item)
		if dataValue.Kind() == reflect.Ptr {
			dataValue = dataValue.Elem()
		}

		idField := dataValue.FieldByName("ID")
		if !idField.IsValid() || idField.IsZero() {
			return nil, errors.New("ID field not found or is zero")
		}

		id, ok := idField.Interface().(string)
		if !ok {
			return nil, errors.New("ID field is not of type string")
		}

		if _, ok := s.data[table]; !ok {
			s.data[table] = make(map[string]interface{})
		}
		s.data[table][id] = item
		ids = append(ids, id)
	}

	return ids, nil
}

func (s *InMemoryStorage) get(table string, id string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tableData, ok := s.data[table]; ok {
		if val, found := tableData[id]; found {
			return val, nil
		}
	}
	return nil, errors.New("item not found")
}

func (s *InMemoryStorage) getRandom(table string, limit int) ([]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []interface{}
	for _, value := range s.data[table] {
		if limit == 0 {
			break
		}
		result = append(result, value)
		limit--
	}
	return result, nil
}

func (s *InMemoryStorage) getAll(table string, matchingFields map[string]string) ([]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []interface{}
	for _, item := range s.data[table] {
		// Check if all matching fields are equal
		match := true
		for field, value := range matchingFields {
			fieldValue := reflect.ValueOf(item).FieldByName(field)
			if !fieldValue.IsValid() {
				return nil, fmt.Errorf("field %s not found", field)
			}
			if fieldValue.Interface() != value {
				match = false
				break
			}
		}

		if match {
			results = append(results, item)
		}
	}
	return results, nil
}

func (s *InMemoryStorage) update(table string, id string, updateFields map[string]interface{}) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Retrieve the existing data for the given table and ID
	tableData, ok := s.data[table]
	if !ok {
		return nil, errors.New("table not found")
	}

	item, found := tableData[id]
	if !found {
		return nil, errors.New("item not found")
	}

	// Use reflection to access and update the fields
	itemValue := reflect.ValueOf(item)
	itemType := reflect.TypeOf(item)

	if itemValue.Kind() == reflect.Ptr {
		// Handle pointers
		if itemValue.IsNil() {
			return nil, errors.New("item pointer is nil")
		}
		itemValue = itemValue.Elem()
		itemType = itemValue.Type()
	}

	// Create a new value of the same type as the item
	if itemValue.Kind() != reflect.Struct {
		return nil, errors.New("item must be a struct or pointer to a struct")
	}

	updatedItem := reflect.New(itemType).Elem()
	updatedItem.Set(itemValue)

	// Update fields in the new value
	for field, newValue := range updateFields {
		fieldValue := updatedItem.FieldByName(field)
		if !fieldValue.IsValid() {
			return nil, fmt.Errorf("field %s not found", field)
		}
		if !fieldValue.CanSet() {
			return nil, fmt.Errorf("field %s cannot be set", field)
		}

		newValueReflect := reflect.ValueOf(newValue)
		if newValueReflect.Type() != fieldValue.Type() {
			return nil, fmt.Errorf("type mismatch for field %s", field)
		}

		fieldValue.Set(newValueReflect)
	}

	// Update the item in the storage
	s.data[table][id] = updatedItem.Interface()
	return updatedItem.Interface(), nil
}

// Helper function for generating unique IDs (simplified)
func getUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
