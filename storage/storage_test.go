package storage

import (
	"testing"
)

// Sample type to store in-memory
func TestStoreAndGet(t *testing.T) {
	storage := NewInMemoryStorage()

	// Store a new template
	template := Template{ID: "1", Title: "Test Template"}
	err := Store(storage, template)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Retrieve the template
	result, err := Get[Template](storage, "1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Title != "Test Template" {
		t.Fatalf("expected 'Test Template', got: %s", result.Title)
	}
}

func TestGetRandom(t *testing.T) {
	storage := NewInMemoryStorage()

	// Store multiple templates
	Store(storage, Template{ID: "1", Title: "Template 1"})
	Store(storage, Template{ID: "2", Title: "Template 2"})
	Store(storage, Template{ID: "3", Title: "Template 3"})

	// Retrieve random templates (limit 2)
	results, err := GetRandom[Template](storage, 2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got: %d", len(results))
	}
}

func TestUpdate(t *testing.T) {
	storage := NewInMemoryStorage()

	// Store a template
	template := Template{ID: "1", Title: "Old Title"}
	Store(storage, template)

	// Update the template
	updateFields := map[string]interface{}{
		"Title": "New Title",
	}
	err := Update[Template](storage, "1", updateFields)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Retrieve and check updated values
	result, err := Get[Template](storage, "1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Title != "New Title" {
		t.Fatalf("expected 'New Title', got: %s", result.Title)
	}
}

func TestStoreAll(t *testing.T) {
	storage := NewInMemoryStorage()

	// Store multiple templates
	templates := []Template{
		{ID: "1", Title: "Template 1"},
		{ID: "2", Title: "Template 2"},
		{ID: "3", Title: "Template 3"},
	}
	err := StoreAll(storage, templates)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Retrieve and check all templates
	result1, err := Get[Template](storage, "1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result1.Title != "Template 1" {
		t.Fatalf("expected 'Template 1', got: %s", result1.Title)
	}

	result2, err := Get[Template](storage, "2")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result2.Title != "Template 2" {
		t.Fatalf("expected 'Template 2', got: %s", result2.Title)
	}

	result3, err := Get[Template](storage, "3")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result3.Title != "Template 3" {
		t.Fatalf("expected 'Template 3', got: %s", result3.Title)
	}
}

type UnregisteredType struct {
	ID    string
	Value string
}

func TestStoreUnregisteredType(t *testing.T) {
	storage := NewInMemoryStorage()

	// Attempt to store an unregistered type
	item := UnregisteredType{ID: "1", Value: "Some Value"}
	err := Store(storage, item)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "table not found for type storage.UnregisteredType" {
		t.Fatalf("expected 'table not found for type storage.UnregisteredType', got: %v", err)
	}
}

func TestStoreAllUnregisteredType(t *testing.T) {
	storage := NewInMemoryStorage()

	// Attempt to store a slice of unregistered types
	items := []UnregisteredType{
		{ID: "1", Value: "Value 1"},
		{ID: "2", Value: "Value 2"},
	}
	err := StoreAll(storage, items)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "table not found for type storage.UnregisteredType" {
		t.Fatalf("expected 'table not found for type storage.UnregisteredType', got: %v", err)
	}
}

func TestGetUnregisteredType(t *testing.T) {
	storage := NewInMemoryStorage()

	// Attempt to get an item of an unregistered type
	var result *UnregisteredType
	result, err := Get[UnregisteredType](storage, "1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "table not found for type storage.UnregisteredType" {
		t.Fatalf("expected 'table not found for type storage.UnregisteredType', got: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}
}

func TestUpdateUnregisteredType(t *testing.T) {
	storage := NewInMemoryStorage()

	// Attempt to update an item of an unregistered type
	updateFields := map[string]interface{}{
		"Value": "New Value",
	}
	err := Update[UnregisteredType](storage, "1", updateFields)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "table not found for type storage.UnregisteredType" {
		t.Fatalf("expected 'table not found for type storage.UnregisteredType', got: %v", err)
	}
}

func TestGetRandomUnregisteredType(t *testing.T) {
	storage := NewInMemoryStorage()

	// Attempt to retrieve random items of an unregistered type
	results, err := GetRandom[UnregisteredType](storage, 2)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != "table not found for type storage.UnregisteredType" {
		t.Fatalf("expected 'table not found for type storage.UnregisteredType', got: %v", err)
	}
	if results != nil {
		t.Fatalf("expected nil results, got: %v", results)
	}
}
