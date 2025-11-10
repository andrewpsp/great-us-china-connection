package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Conn3ction DNS API Server...")

	// Initialize store
	store, err := InitStore()
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Close etcd connection if applicable
	if etcdStore, ok := store.(*EtcdStore); ok {
		defer etcdStore.Close()
	}

	// Create API handler
	api := NewAPI(store)

	// Setup router
	router := mux.NewRouter()
	router.HandleFunc("/healthz", api.HealthzHandler).Methods("GET")
	router.HandleFunc("/records", api.ListRecordsHandler).Methods("GET")
	router.HandleFunc("/records", api.CreateRecordHandler).Methods("POST")
	router.HandleFunc("/records/{name}", api.GetRecordHandler).Methods("GET")
	router.HandleFunc("/records/{name}", api.UpdateRecordHandler).Methods("PUT")
	router.HandleFunc("/records/{name}", api.DeleteRecordHandler).Methods("DELETE")

	// Setup HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
