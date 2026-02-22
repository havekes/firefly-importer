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

	// Poll the server until it responds or we time out (max 500ms)
	var (
		resp *http.Response
		err  error
	)
	for i := 0; i < 10; i++ {
		time.Sleep(50 * time.Millisecond)
		resp, err = http.Get("http://localhost:8089/")
		if err == nil {
			break
		}
	}
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
	// IndexHandler calls the Firefly API which is unreachable → expect 500
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("GET / expected 500 (no Firefly backend), got %d", rr.Code)
	}

	// Test POST /upload route
	// UploadHandler: no multipart body → 400 Bad Request
	req, _ = http.NewRequest("POST", "/upload", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("POST /upload expected 400, got %d", rr.Code)
	}

	// Test POST /save route
	// SaveHandler: empty transactions list → 200 OK with added:0
	req, _ = http.NewRequest("POST", "/save", strings.NewReader(`{"transactions":[]}`))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /save expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"added":0`) {
		t.Errorf("POST /save unexpected body: %s", rr.Body.String())
	}
}
