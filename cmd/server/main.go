// File: cmd/server/main.go

package main

import (
	"fmt"
	"log"
	"net/http"

	"bp-tracker/internal/handlers"
	"bp-tracker/internal/database"
)

func main() {
	// Initialize database
	db, err := database.New("bp.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	handler, err := handlers.New(db)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	// Get routes with middleware applied
	mux := handler.Routes()

	// Start server
	addr := fmt.Sprintf(":%d", 32401)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
