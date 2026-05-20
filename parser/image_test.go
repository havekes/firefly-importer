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

	txs, err := ParseImage(imageReader, "2023-11-15", mockServer.URL, "test-key", "gpt-4-vision-preview")
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

func TestCleanVisionResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no thinking tokens or markdown",
			input:    `[{"date":"2026-05-16"}]`,
			expected: `[{"date":"2026-05-16"}]`,
		},
		{
			name:     "with thought block",
			input:    `<thought>The user wants transactions.</thought>[{"date":"2026-05-16"}]`,
			expected: `[{"date":"2026-05-16"}]`,
		},
		{
			name:     "with think block",
			input:    `<think>Evaluating image...</think>[{"date":"2026-05-16"}]`,
			expected: `[{"date":"2026-05-16"}]`,
		},
		{
			name:     "case-insensitive thought block",
			input:    `<THOUGHT>Thinking...</tHoUgHt>[{"date":"2026-05-16"}]`,
			expected: `[{"date":"2026-05-16"}]`,
		},
		{
			name:     "case-insensitive think block",
			input:    `<THINK>Thinking...</tHiNk>[{"date":"2026-05-16"}]`,
			expected: `[{"date":"2026-05-16"}]`,
		},
		{
			name:     "unclosed thought block",
			input:    `<thought>Thinking...`,
			expected: ``,
		},
		{
			name:     "unclosed think block",
			input:    `<think>Thinking...`,
			expected: ``,
		},
		{
			name:     "with markdown code blocks",
			input:    "```json\n[{\"date\":\"2026-05-16\"}]\n```",
			expected: `[{"date":"2026-05-16"}]`,
		},
		{
			name:     "thought block and markdown code blocks",
			input:    "<thought>thinking</thought>\n```json\n[{\"date\":\"2026-05-16\"}]\n```",
			expected: `[{"date":"2026-05-16"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanVisionResponse(tt.input)
			if got != tt.expected {
				t.Errorf("cleanVisionResponse(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
