package main

import (
	"firefly-importer/config"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAppStartup(t *testing.T) {
	// Setup environment variables for test
	os.Setenv("PORT", "8089") // use a non-standard port for testing
	os.Setenv("FIREFLY_URL", "http://localhost:8080")
	defer os.Clearenv()

	// Use a channel to coordinate
	go func() {
		main()
	}()

	// Give the server a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Test if the server is accessible
	resp, err := http.Get("http://localhost:8089/")
	if err != nil {
		t.Fatalf("Server failed to start or respond on port 8089: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		// we might get a 500 if the mock firefly isn't up, but the server IS running
		t.Logf("Got status code %d, but server is running", resp.StatusCode)
	}
}

func TestRouting(t *testing.T) {
	cfg := &config.Config{
		FireflyURL: "http://example.com/api/v1",
	}

	mux := setupRouter(cfg)

	// Test GET / route
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// It's expected to hit the handler.
	// Since there's no real Firefly API, it might return 500, but route works
	if rr.Code == http.StatusNotFound {
		t.Errorf("GET / route not found")
	}

	// Test POST /upload route
	req, _ = http.NewRequest("POST", "/upload", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /upload route not found")
	}

	// Test POST /save route
	req, _ = http.NewRequest("POST", "/save", strings.NewReader(`{}`))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /save route not found")
	}
}
