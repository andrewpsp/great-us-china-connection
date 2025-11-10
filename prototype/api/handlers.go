package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ListRecordsHandler returns all records
func ListRecordsHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := store.List()
		if err != nil {
			log.Printf("error listing records: %v", err)
			http.Error(w, "failed to list records", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	}
}

// GetRecordHandler retrieves a specific record by name
func GetRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		record, err := store.Get(name)
		if err != nil {
			log.Printf("error getting record %s: %v", name, err)
			http.Error(w, "record not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}
}

// CreateRecordHandler creates a new record
func CreateRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var record Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			log.Printf("error decoding record: %v", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if record.Name == "" || record.Type == "" || len(record.Values) == 0 {
			http.Error(w, "name, type, and values are required", http.StatusBadRequest)
			return
		}

		if err := store.Put(record); err != nil {
			log.Printf("error creating record: %v", err)
			http.Error(w, "failed to create record", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)
	}
}

// UpdateRecordHandler updates an existing record
func UpdateRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		var record Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			log.Printf("error decoding record: %v", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Ensure the name matches the URL parameter
		record.Name = name

		if record.Type == "" || len(record.Values) == 0 {
			http.Error(w, "type and values are required", http.StatusBadRequest)
			return
		}

		if err := store.Put(record); err != nil {
			log.Printf("error updating record: %v", err)
			http.Error(w, "failed to update record", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}
}

// DeleteRecordHandler deletes a record
func DeleteRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		if err := store.Delete(name); err != nil {
			log.Printf("error deleting record %s: %v", name, err)
			http.Error(w, "record not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
