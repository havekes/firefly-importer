package parser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseImage(t *testing.T) {
	// Create a mock Vision API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		mockResponse := visionResponse{}
		responseContent := `[{"date":"2023-11-15","description":"Coffee Shop","amount":4.50,"type":"withdrawal"}]`

		mockResponse.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: responseContent,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Provide a dummy image
	imageReader := strings.NewReader("dummy image content representing bytes")

	txs, err := ParseImage(imageReader, mockServer.URL, "test-key", "gpt-4-vision-preview")
	if err != nil {
		t.Fatalf("ParseImage failed: %v", err)
	}

	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	if txs[0].Date != "2023-11-15" {
		t.Errorf("Expected Date 2023-11-15, got %s", txs[0].Date)
	}
	if txs[0].Description != "Coffee Shop" {
		t.Errorf("Expected Description Coffee Shop, got %s", txs[0].Description)
	}
	if txs[0].Amount != 4.50 {
		t.Errorf("Expected Amount 4.50, got %f", txs[0].Amount)
	}
	if txs[0].Type != "withdrawal" {
		t.Errorf("Expected Type withdrawal, got %s", txs[0].Type)
	}
}
