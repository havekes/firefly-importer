package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"firefly-importer/config"
	"firefly-importer/db"
	"firefly-importer/firefly"
	"firefly-importer/handlers"

	"github.com/gorilla/csrf"
)

// setupRouter configures the dependencies and routes
func setupRouter(cfg *config.Config, dbConn *sql.DB) *http.ServeMux {
	client := firefly.NewClient(cfg.FireflyURL, cfg.FireflyToken)
	appHandler := handlers.NewAppHandler(client, cfg, dbConn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", appHandler.IndexHandler)
	mux.HandleFunc("POST /upload", appHandler.UploadHandler)
	mux.HandleFunc("POST /save", appHandler.SaveHandler)

	return mux
}

// plaintextMiddleware marks every request as plaintext HTTP so that
// gorilla/csrf skips the strict HTTPS-only Referer validation.
func plaintextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), csrf.PlaintextHTTPContextKey, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	cfg := config.LoadConfig()

	if cfg.Debug {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
		slog.SetDefault(logger)

		db.EnableQueryLogging(true)
		log.Println("DEBUG mode and query logging enabled")
	}

	log.Printf("Starting Firefly Importer on port %s", cfg.Port)

	dbConn, err := db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Printf("Failed to initialize database (name mappings disabled): %v", err)
	} else {
		defer dbConn.Close()
	}

	mux := setupRouter(cfg, dbConn)

	// Derive a 32-byte CSRF key from config
	csrfKey := []byte(cfg.CSRFKey)
	if len(csrfKey) < 32 {
		padded := make([]byte, 32)
		copy(padded, csrfKey)
		csrfKey = padded
	} else if len(csrfKey) > 32 {
		csrfKey = csrfKey[:32]
	}

	// Wrap the mux with CSRF protection middleware
	// Secure(false) ensures the CSRF cookie is sent over HTTP (not just HTTPS)
	// Path("/") ensures the cookie applies to all routes
	// Build CSRF options
	csrfOpts := []csrf.Option{
		csrf.Secure(false),
		csrf.Path("/"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := csrf.FailureReason(r)
			http.Error(w, fmt.Sprintf("CSRF Validation Failed: %v", err), http.StatusForbidden)
		})),
	}

	// When behind a reverse proxy (HTTPS → HTTP), the browser sends
	// Origin: https://hostname but the app sees plain HTTP. TrustedOrigins
	// tells gorilla/csrf to accept this mismatch.
	if cfg.Hostname != "" {
		log.Printf("Adding TrustedOrigin: %s", cfg.Hostname)
		csrfOpts = append(csrfOpts, csrf.TrustedOrigins([]string{cfg.Hostname}))
	}

	csrfMiddleware := csrf.Protect(csrfKey, csrfOpts...)

	// plaintextMiddleware tells gorilla/csrf we're serving over HTTP,
	// so it skips the strict HTTPS Referer validation that causes 403s.
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	if err := http.ListenAndServe(serverAddr, plaintextMiddleware(csrfMiddleware(mux))); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
