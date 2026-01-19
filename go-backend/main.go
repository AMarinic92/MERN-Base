package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"go-backend/database"
	"go-backend/handlers"
	"go-backend/models"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// 1. Initialize the database connection and run migrations
	database.InitializeDatabase(&models.Card{})
	database.InitializeMemgraph()
	// 2. Check if we should prime the database in the background
	if len(os.Args) > 1 && os.Args[1] == "prime" {
		filePath := "../../all-cards.json"
		if len(os.Args) > 2 {
			filePath = os.Args[2]
		}

		// Start priming in a goroutine
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			fmt.Println("Starting database priming in background...")

			file, err := os.Open(filePath)
			if err != nil {
				log.Printf("Failed to open JSON file '%s': %v", filePath, err)
				return
			}
			defer file.Close()

			if err := database.PrimeDatabase(file); err != nil {
				log.Printf("Failed to prime database: %v", err)
				return
			}

			fmt.Println("Database primed successfully!")
		}()

	}

	// 3. Setup the router
	router := mux.NewRouter()

	// 4. Define the routes
	router.HandleFunc("/api/cards/rand", handlers.GetRndCard).Methods("GET")
	router.HandleFunc("/api/cards/similar", handlers.GetSimilarCards).Methods("POST")
	router.HandleFunc("/api/cards/fuzzy",handlers.GetFuzzyCard).Methods("GET")
	router.HandleFunc("/api/cards/id",handlers.GetCardID).Methods("GET")
	router.HandleFunc("/api/cards/mems", handlers.MemSuggest).Methods("POST")

	router.PathPrefix("/").HandlerFunc(handlers.OptionsHandler).Methods("OPTIONS")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // For development - allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true, // Enable debug mode to see CORS logs
	})

	// Wrap router with CORS middleware
	handler := c.Handler(router)

	// 5. Start the server
	port := "8081"
	fmt.Printf("Server listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
