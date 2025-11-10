package main

import (
	"testing"
)

func TestMemStore(t *testing.T) {
	store := NewMemStore()

	// Test Put and Get
	record := Record{
		Name:   "test.example.local",
		Type:   "A",
		Values: []string{"10.1.2.3"},
		TTL:    60,
	}

	err := store.Put(record)
	if err != nil {
		t.Fatalf("failed to put record: %v", err)
	}

	retrieved, err := store.Get("test.example.local")
	if err != nil {
		t.Fatalf("failed to get record: %v", err)
	}

	if retrieved.Name != record.Name {
		t.Errorf("expected name %s, got %s", record.Name, retrieved.Name)
	}

	// Test List
	records, err := store.List()
	if err != nil {
		t.Fatalf("failed to list records: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}

	// Test Update
	record.Values = []string{"10.1.2.4"}
	err = store.Put(record)
	if err != nil {
		t.Fatalf("failed to update record: %v", err)
	}

	retrieved, err = store.Get("test.example.local")
	if err != nil {
		t.Fatalf("failed to get updated record: %v", err)
	}

	if len(retrieved.Values) != 1 || retrieved.Values[0] != "10.1.2.4" {
		t.Errorf("expected value 10.1.2.4, got %v", retrieved.Values)
	}

	// Test Delete
	err = store.Delete("test.example.local")
	if err != nil {
		t.Fatalf("failed to delete record: %v", err)
	}

	_, err = store.Get("test.example.local")
	if err == nil {
		t.Error("expected error when getting deleted record")
	}

	// Test Delete non-existent
	err = store.Delete("nonexistent.example.local")
	if err == nil {
		t.Error("expected error when deleting non-existent record")
	}
}

func TestMemStoreConcurrent(t *testing.T) {
	store := NewMemStore()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			record := Record{
				Name:   "concurrent.example.local",
				Type:   "A",
				Values: []string{"10.1.2.3"},
				TTL:    60,
			}
			store.Put(record)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify record exists
	_, err := store.Get("concurrent.example.local")
	if err != nil {
		t.Fatalf("failed to get concurrent record: %v", err)
	}
}
