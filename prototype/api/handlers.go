package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// API handles HTTP requests
type API struct {
	store Store
}

// NewAPI creates a new API handler
func NewAPI(store Store) *API {
	return &API{store: store}
}

// HealthzHandler handles health check requests
func (a *API) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ListRecordsHandler handles GET /records
func (a *API) ListRecordsHandler(w http.ResponseWriter, r *http.Request) {
	records, err := a.store.List()
	if err != nil {
		log.Printf("Error listing records: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// GetRecordHandler handles GET /records/{name}
func (a *API) GetRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	record, found, err := a.store.Get(name)
	if err != nil {
		log.Printf("Error getting record: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// CreateRecordHandler handles POST /records
func (a *API) CreateRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate record
	if record.Name == "" {
		http.Error(w, "Record name is required", http.StatusBadRequest)
		return
	}
	if record.Type == "" {
		record.Type = "A"
	}
	if len(record.Values) == 0 {
		http.Error(w, "Record values are required", http.StatusBadRequest)
		return
	}
	if record.TTL <= 0 {
		record.TTL = 60
	}

	if err := a.store.Put(record); err != nil {
		log.Printf("Error creating record: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// UpdateRecordHandler handles PUT /records/{name}
func (a *API) UpdateRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var record Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set name from URL if not provided in body
	if record.Name == "" {
		record.Name = name
	} else if record.Name != name {
		http.Error(w, "Record name in URL and body must match", http.StatusBadRequest)
		return
	}

	// Validate record
	if record.Type == "" {
		record.Type = "A"
	}
	if len(record.Values) == 0 {
		http.Error(w, "Record values are required", http.StatusBadRequest)
		return
	}
	if record.TTL <= 0 {
		record.TTL = 60
	}

	if err := a.store.Put(record); err != nil {
		log.Printf("Error updating record: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// DeleteRecordHandler handles DELETE /records/{name}
func (a *API) DeleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := a.store.Delete(name); err != nil {
		log.Printf("Error deleting record: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
