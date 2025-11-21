package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"go-backend/database"
	"go-backend/handlers"
	"go-backend/models"
)

func main() {
	// 1. Initialize the database connection and run migrations
	// Ensure all models are passed to AutoMigrate
	database.InitializeDatabase(&models.Product{})

	// 2. Setup the router
	router := mux.NewRouter()

	// 3. Define the routes, linking to handlers
	router.HandleFunc("/products", handlers.GetProducts).Methods("GET")
	router.HandleFunc("/products", handlers.CreateProduct).Methods("POST")
	router.HandleFunc("/products/{id}", handlers.DeleteProduct).Methods("DELETE")

	// Add OPTIONS handler for all paths to handle CORS preflight
	router.PathPrefix("/").HandlerFunc(handlers.OptionsHandler).Methods("OPTIONS")

	// 4. Start the server
	port := "8080"
	fmt.Printf("Server listening on http://localhost:%s\n", port)

	// Create necessary directories if needed (for SQLite in this case)
	if err := os.MkdirAll("./", 0755); err != nil {
		log.Fatalf("Could not create database directory: %v", err)
	}

	log.Fatal(http.ListenAndServe(":"+port, router))
}
