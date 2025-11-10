package main

import (
	"testing"
)

func TestMemStore(t *testing.T) {
	store := NewMemStore()

	// Test Put and Get
	record := Record{
		Name:   "test.example.com",
		Type:   "A",
		Values: []string{"1.2.3.4"},
		TTL:    60,
	}

	err := store.Put(record)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, exists, err := store.Get("test.example.com")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !exists {
		t.Fatal("Record not found after Put")
	}
	if retrieved.Name != record.Name {
		t.Errorf("Expected name %s, got %s", record.Name, retrieved.Name)
	}

	// Test List
	records, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Test Delete
	err = store.Delete("test.example.com")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, exists, err = store.Get("test.example.com")
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if exists {
		t.Error("Record still exists after Delete")
	}
}

func TestMemStoreList(t *testing.T) {
	store := NewMemStore()

	records := []Record{
		{Name: "app1.example.com", Type: "A", Values: []string{"1.2.3.4"}, TTL: 60},
		{Name: "app2.example.com", Type: "A", Values: []string{"5.6.7.8"}, TTL: 60},
		{Name: "app3.example.com", Type: "A", Values: []string{"9.10.11.12"}, TTL: 60},
	}

	for _, record := range records {
		if err := store.Put(record); err != nil {
			t.Fatalf("Put failed: %v", err)
		}
	}

	retrieved, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(retrieved) != len(records) {
		t.Errorf("Expected %d records, got %d", len(records), len(retrieved))
	}
}

func TestInitStore(t *testing.T) {
	// Test in-memory store initialization (no ETCD_ENDPOINTS set)
	store, err := InitStore()
	if err != nil {
		t.Fatalf("InitStore failed: %v", err)
	}

	if _, ok := store.(*MemStore); !ok {
		t.Error("Expected MemStore when ETCD_ENDPOINTS is not set")
	}
}
