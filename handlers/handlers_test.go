package handlers

import (
	"firefly-importer/config"
	"firefly-importer/firefly"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexHandler(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := `{
			"data": [
				{
					"id": "1",
					"attributes": {
						"name": "Checking Account",
						"type": "asset"
					}
				}
			]
		}`
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := firefly.NewClient(mockServer.URL, "test-token")
	cfg := &config.Config{}
	appHandler := NewAppHandler(client, cfg)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	appHandler.IndexHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Firefly III Statement Importer") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Checking Account") {
		t.Errorf("handler didn't include accounts in body: got %v", rr.Body.String())
	}
}

func TestSaveHandler(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := firefly.NewClient(mockServer.URL, "test-token")
	cfg := &config.Config{}
	appHandler := NewAppHandler(client, cfg)

	body := `{"transactions": [{"date": "2023-12-01", "description": "Test", "amount": 10.0, "type": "withdrawal", "source_id": "1", "status": "Added"}]}`
	req, err := http.NewRequest("POST", "/save", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	appHandler.SaveHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), `"added":1`) {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}
