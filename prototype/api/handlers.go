package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func setupRouter(store Store) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthz", healthzHandler).Methods("GET")
	router.HandleFunc("/records", listRecordsHandler(store)).Methods("GET")
	router.HandleFunc("/records", createRecordHandler(store)).Methods("POST")
	router.HandleFunc("/records/{name}", updateRecordHandler(store)).Methods("PUT")
	router.HandleFunc("/records/{name}", deleteRecordHandler(store)).Methods("DELETE")
	router.HandleFunc("/records/{name}", getRecordHandler(store)).Methods("GET")

	return router
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func listRecordsHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := store.List()
		if err != nil {
			log.Printf("Error listing records: %v", err)
			http.Error(w, "Failed to list records", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": records,
		})
	}
}

func createRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var record Record
		if err := json.Unmarshal(body, &record); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if record.Name == "" || record.Type == "" || len(record.Values) == 0 {
			http.Error(w, "Missing required fields: name, type, values", http.StatusBadRequest)
			return
		}

		if err := store.Put(record); err != nil {
			log.Printf("Error creating record: %v", err)
			http.Error(w, "Failed to create record", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)
	}
}

func updateRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var record Record
		if err := json.Unmarshal(body, &record); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Ensure the name matches the URL
		record.Name = name

		if record.Type == "" || len(record.Values) == 0 {
			http.Error(w, "Missing required fields: type, values", http.StatusBadRequest)
			return
		}

		if err := store.Put(record); err != nil {
			log.Printf("Error updating record: %v", err)
			http.Error(w, "Failed to update record", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}
}

func deleteRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		if err := store.Delete(name); err != nil {
			log.Printf("Error deleting record: %v", err)
			http.Error(w, "Failed to delete record", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getRecordHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		record, exists, err := store.Get(name)
		if err != nil {
			log.Printf("Error getting record: %v", err)
			http.Error(w, "Failed to get record", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}
}
