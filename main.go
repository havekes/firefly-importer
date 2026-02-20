package main

import (
	"firefly-importer/config"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting Firefly Importer on port %s", cfg.Port)
	log.Printf("Using Firefly API URL: %s", cfg.FireflyURL)

	// Set up basic HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the Firefly III Statement Importer!")
	})

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
