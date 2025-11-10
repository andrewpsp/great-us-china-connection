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

// Store is the interface for storing DNS records
type Store interface {
	List() ([]Record, error)
	Get(name string) (Record, bool, error)
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

func (m *MemStore) List() ([]Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]Record, 0, len(m.records))
	for _, r := range m.records {
		records = append(records, r)
	}
	return records, nil
}

func (m *MemStore) Get(name string) (Record, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.records[name]
	return record, ok, nil
}

func (m *MemStore) Put(record Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.records[record.Name] = record
	return nil
}

func (m *MemStore) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdStore{
		client: client,
		prefix: prefix,
	}, nil
}

func (e *EtcdStore) Close() error {
	return e.client.Close()
}

func (e *EtcdStore) key(name string) string {
	return fmt.Sprintf("%s/%s", e.prefix, name)
}

func (e *EtcdStore) List() ([]Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, e.prefix+"/", clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	records := make([]Record, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var record Record
		if err := json.Unmarshal(kv.Value, &record); err != nil {
			log.Printf("failed to unmarshal record: %v", err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

func (e *EtcdStore) Get(name string) (Record, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, e.key(name))
	if err != nil {
		return Record{}, false, fmt.Errorf("failed to get record: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return Record{}, false, nil
	}

	var record Record
	if err := json.Unmarshal(resp.Kvs[0].Value, &record); err != nil {
		return Record{}, false, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return record, true, nil
}

func (e *EtcdStore) Put(record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	_, err = e.client.Put(ctx, e.key(record.Name), string(data))
	if err != nil {
		return fmt.Errorf("failed to put record: %w", err)
	}

	return nil
}

func (e *EtcdStore) Delete(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.client.Delete(ctx, e.key(name))
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// InitStore initializes the appropriate store based on environment variables
func InitStore() (Store, error) {
	etcdEndpoints := os.Getenv("ETCD_ENDPOINTS")
	if etcdEndpoints == "" {
		log.Println("ETCD_ENDPOINTS not set, using in-memory store")
		return NewMemStore(), nil
	}

	endpoints := strings.Split(etcdEndpoints, ",")
	for i, ep := range endpoints {
		endpoints[i] = strings.TrimSpace(ep)
	}

	prefix := os.Getenv("ETCD_PREFIX")
	if prefix == "" {
		prefix = "/conn3ction/records"
	}

	log.Printf("Initializing etcd store with endpoints: %v, prefix: %s", endpoints, prefix)
	store, err := NewEtcdStore(endpoints, prefix)
	if err != nil {
		log.Printf("Failed to connect to etcd: %v, falling back to in-memory store", err)
		return NewMemStore(), nil
	}

	return store, nil
}
