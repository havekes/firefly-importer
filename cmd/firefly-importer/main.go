package main

import (
	"fmt"
	"log"
	"net/http"

	"firefly-importer/config"
	"firefly-importer/firefly"
	"firefly-importer/handlers"
)

// setupRouter configures the dependencies and routes
func setupRouter(cfg *config.Config) *http.ServeMux {
	client := firefly.NewClient(cfg.FireflyURL, cfg.FireflyToken)
	appHandler := handlers.NewAppHandler(client, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", appHandler.IndexHandler)
	mux.HandleFunc("POST /upload", appHandler.UploadHandler)
	mux.HandleFunc("POST /save", appHandler.SaveHandler)

	return mux
}

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting Firefly Importer on port %s", cfg.Port)
	log.Printf("Using Firefly API URL: %s", cfg.FireflyURL)

	mux := setupRouter(cfg)

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
