package firefly

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAccounts(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/accounts" {
			t.Errorf("Expected path /accounts, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "asset" {
			t.Errorf("Expected type=asset, got %s", r.URL.Query().Get("type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}

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

	client := NewClient(mockServer.URL, "test-token")
	accounts, err := client.GetAccounts()

	if err != nil {
		t.Fatalf("GetAccounts failed: %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(accounts))
	}

	if accounts[0].ID != "1" {
		t.Errorf("Expected ID 1, got %s", accounts[0].ID)
	}
	if accounts[0].Name != "Checking Account" {
		t.Errorf("Expected Name Checking Account, got %s", accounts[0].Name)
	}
}
