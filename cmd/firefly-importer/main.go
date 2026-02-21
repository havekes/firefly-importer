package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"firefly-importer/config"
	"firefly-importer/firefly"
	"firefly-importer/handlers"

	"github.com/gorilla/csrf"
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

	// Generate a 32-byte random key for CSRF
	csrfKey := make([]byte, 32)
	if _, err := rand.Read(csrfKey); err != nil {
		log.Fatalf("Failed to generate CSRF key: %v", err)
	}

	// Wrap the mux with CSRF protection middleware
	// Setting Secure(false) since local development typically runs over HTTP
	csrfMiddleware := csrf.Protect(csrfKey, csrf.Secure(false))

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	if err := http.ListenAndServe(serverAddr, csrfMiddleware(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
