package main

import (
	"context"
	"encoding/json"
	"fmt"
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

// Store defines the interface for record storage
type Store interface {
	List() ([]Record, error)
	Get(name string) (*Record, error)
	Put(record Record) error
	Delete(name string) error
}

// MemStore is an in-memory implementation of Store
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
func (m *MemStore) Get(name string) (*Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, exists := m.records[name]
	if !exists {
		return nil, fmt.Errorf("record not found: %s", name)
	}
	return &record, nil
}

// Put creates or updates a record
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

	if _, exists := m.records[name]; !exists {
		return fmt.Errorf("record not found: %s", name)
	}
	delete(m.records, name)
	return nil
}

// EtcdStore is an etcd-backed implementation of Store
type EtcdStore struct {
	client *clientv3.Client
	prefix string
}

// NewEtcdStore creates a new etcd-backed store
func NewEtcdStore(endpoints []string, prefix string) (*EtcdStore, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	return &EtcdStore{
		client: cli,
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
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	records := make([]Record, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var record Record
		if err := json.Unmarshal(kv.Value, &record); err != nil {
			log.Printf("warning: failed to unmarshal record %s: %v", kv.Key, err)
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// Get retrieves a record by name from etcd
func (e *EtcdStore) Get(name string) (*Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := e.prefix + "/" + name
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("record not found: %s", name)
	}

	var record Record
	if err := json.Unmarshal(resp.Kvs[0].Value, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}
	return &record, nil
}

// Put creates or updates a record in etcd
func (e *EtcdStore) Put(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	key := e.prefix + "/" + record.Name
	_, err = e.client.Put(ctx, key, string(data))
	if err != nil {
		return fmt.Errorf("failed to put record: %w", err)
	}
	return nil
}

// Delete removes a record from etcd
func (e *EtcdStore) Delete(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := e.prefix + "/" + name
	resp, err := e.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	if resp.Deleted == 0 {
		return fmt.Errorf("record not found: %s", name)
	}
	return nil
}

// NewStore creates a store based on environment configuration
func NewStore() (Store, error) {
	etcdEndpoints := os.Getenv("ETCD_ENDPOINTS")
	if etcdEndpoints == "" {
		log.Println("ETCD_ENDPOINTS not set, using in-memory store")
		return NewMemStore(), nil
	}

	endpoints := strings.Split(etcdEndpoints, ",")
	for i := range endpoints {
		endpoints[i] = strings.TrimSpace(endpoints[i])
	}

	prefix := os.Getenv("ETCD_PREFIX")
	if prefix == "" {
		prefix = "/polycloud/records"
	}

	log.Printf("Connecting to etcd at %v with prefix %s", endpoints, prefix)
	return NewEtcdStore(endpoints, prefix)
}
