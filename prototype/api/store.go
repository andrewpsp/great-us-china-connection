package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Record represents a DNS record
type Record struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []string `json:"values"`
	TTL    int      `json:"ttl"`
}

// Store interface for record storage
type Store interface {
	List() ([]Record, error)
	Get(name string) (Record, bool, error)
	Put(record Record) error
	Delete(name string) error
}

// MemStore is an in-memory store implementation
type MemStore struct {
	mu      sync.RWMutex
	records map[string]Record
}

// NewMemStore creates a new in-memory store
func NewMemStore() *MemStore {
	return &MemStore{
		records: make(map[string]Record),
	}
}

// List returns all records
func (m *MemStore) List() ([]Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]Record, 0, len(m.records))
	for _, record := range m.records {
		records = append(records, record)
	}
	return records, nil
}

// Get retrieves a record by name
func (m *MemStore) Get(name string) (Record, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, exists := m.records[name]
	return record, exists, nil
}

// Put stores or updates a record
func (m *MemStore) Put(record Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.records[record.Name] = record
	return nil
}

// Delete removes a record by name
func (m *MemStore) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.records, name)
	return nil
}

// EtcdStore is an etcd-backed store implementation
type EtcdStore struct {
	client *clientv3.Client
	prefix string
}

// NewEtcdStore creates a new etcd-backed store
func NewEtcdStore(endpoints []string, prefix string) (*EtcdStore, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdStore{
		client: client,
		prefix: prefix,
	}, nil
}

// Close closes the etcd client connection
func (e *EtcdStore) Close() error {
	return e.client.Close()
}

// List returns all records from etcd
func (e *EtcdStore) List() ([]Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, e.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	records := make([]Record, 0)
	for _, kv := range resp.Kvs {
		var record Record
		if err := unmarshalRecord(kv.Value, &record); err != nil {
			log.Printf("Error unmarshaling record: %v", err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// Get retrieves a record by name from etcd
func (e *EtcdStore) Get(name string) (Record, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := e.prefix + name
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return Record{}, false, err
	}

	if len(resp.Kvs) == 0 {
		return Record{}, false, nil
	}

	var record Record
	if err := unmarshalRecord(resp.Kvs[0].Value, &record); err != nil {
		return Record{}, false, err
	}

	return record, true, nil
}

// Put stores or updates a record in etcd
func (e *EtcdStore) Put(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := e.prefix + record.Name
	value, err := marshalRecord(record)
	if err != nil {
		return err
	}

	_, err = e.client.Put(ctx, key, value)
	return err
}

// Delete removes a record by name from etcd
func (e *EtcdStore) Delete(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := e.prefix + name
	_, err := e.client.Delete(ctx, key)
	return err
}

// Helper functions for JSON marshaling
func marshalRecord(record Record) (string, error) {
	// Use encoding/json for proper marshaling
	data, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalRecord(data []byte, record *Record) error {
	return json.Unmarshal(data, record)
}

// InitStore initializes the appropriate store based on environment variables
func InitStore() (Store, error) {
	etcdEndpoints := os.Getenv("ETCD_ENDPOINTS")
	if etcdEndpoints == "" {
		log.Println("ETCD_ENDPOINTS not set, using in-memory store")
		return NewMemStore(), nil
	}

	endpoints := strings.Split(etcdEndpoints, ",")
	prefix := os.Getenv("ETCD_PREFIX")
	if prefix == "" {
		prefix = "/conn3ction/records/"
	}

	log.Printf("Connecting to etcd at %v with prefix %s", endpoints, prefix)
	store, err := NewEtcdStore(endpoints, prefix)
	if err != nil {
		log.Printf("Failed to connect to etcd: %v, falling back to in-memory store", err)
		return NewMemStore(), nil
	}

	return store, nil
}
