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
	// Initialize store
	store, err := NewStore()
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	// Close etcd connection if using EtcdStore
	if etcdStore, ok := store.(*EtcdStore); ok {
		defer etcdStore.Close()
	}

	// Setup router
	router := mux.NewRouter()
	router.HandleFunc("/healthz", HealthHandler).Methods("GET")
	router.HandleFunc("/records", ListRecordsHandler(store)).Methods("GET")
	router.HandleFunc("/records", CreateRecordHandler(store)).Methods("POST")
	router.HandleFunc("/records/{name}", GetRecordHandler(store)).Methods("GET")
	router.HandleFunc("/records/{name}", UpdateRecordHandler(store)).Methods("PUT")
	router.HandleFunc("/records/{name}", DeleteRecordHandler(store)).Methods("DELETE")

	// Setup HTTP server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Starting PolyCloud DNS API server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed to start: %v", err)
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
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
