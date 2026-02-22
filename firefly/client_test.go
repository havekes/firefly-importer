package firefly

import (
	"encoding/json"
	"firefly-importer/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRecentTransactions(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/accounts/123/transactions" {
			t.Errorf("Expected path /accounts/123/transactions, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("start") == "" {
			t.Errorf("Expected start query parameter to be set")
		}
		if r.URL.Query().Get("end") == "" {
			t.Errorf("Expected end query parameter to be set")
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}

		mockResponse := `{
			"data": [
				{
					"attributes": {
						"transactions": [
							{
								"date": "2023-12-01T00:00:00+00:00",
								"description": "Internet Bill",
								"amount": "60.00",
								"type": "withdrawal",
								"source_name": "Checking Account",
								"destination_name": "ISP"
							}
						]
					}
				}
			]
		}`

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, "test-token")
	txs, err := client.GetRecentTransactions("123", 30)

	if err != nil {
		t.Fatalf("GetRecentTransactions failed: %v", err)
	}

	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	if txs[0].Date != "2023-12-01" {
		t.Errorf("Expected extracted date 2023-12-01, got %s", txs[0].Date)
	}
	if txs[0].Amount != 60.00 {
		t.Errorf("Expected amount 60.00, got %f", txs[0].Amount)
	}
	if txs[0].Description != "Internet Bill" {
		t.Errorf("Expected Description Internet Bill, got %s", txs[0].Description)
	}
	if txs[0].SourceName != "Checking Account" {
		t.Errorf("Expected SourceName Checking Account, got %s", txs[0].SourceName)
	}
	if txs[0].DestinationName != "ISP" {
		t.Errorf("Expected DestinationName ISP, got %s", txs[0].DestinationName)
	}
	if txs[0].Status != models.StatusAdded {
		t.Errorf("Expected status %s, got %s", models.StatusAdded, txs[0].Status)
	}
}

func TestStoreTransaction(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/transactions" {
			t.Errorf("Expected path /transactions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}

		var reqPayload fireflyStoreTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
			t.Fatalf("Failed to decode store request: %v", err)
		}

		if len(reqPayload.Transactions) != 1 {
			t.Fatalf("Expected 1 transaction in payload, got %d", len(reqPayload.Transactions))
		}

		tx := reqPayload.Transactions[0]
		if tx.Amount != "12.50" { // Should be formatted to string
			t.Errorf("Expected amount string '12.50', got %s", tx.Amount)
		}
		if tx.SourceName != "Wallet" {
			t.Errorf("Expected SourceName 'Wallet', got %s", tx.SourceName)
		}
		if tx.DestinationName != "Restaurant" {
			t.Errorf("Expected DestinationName 'Restaurant', got %s", tx.DestinationName)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, "test-token")

	newTx := models.Transaction{
		Date:            "2023-12-05",
		Description:     "Lunch",
		Amount:          12.50,
		Type:            "withdrawal",
		SourceName:      "Wallet",
		DestinationName: "Restaurant",
	}

	err := client.StoreTransaction(newTx)
	if err != nil {
		t.Fatalf("StoreTransaction failed: %v", err)
	}
}

func TestStoreTransactions(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var reqPayload fireflyStoreTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
			t.Fatalf("Failed to decode store request: %v", err)
		}

		if len(reqPayload.Transactions) != 2 {
			t.Fatalf("Expected 2 transactions in payload, got %d", len(reqPayload.Transactions))
		}

		if reqPayload.Transactions[0].Description != "Lunch" || reqPayload.Transactions[1].Description != "Dinner" {
			t.Errorf("Unexpected transaction descriptions")
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, "test-token")

	txs := []models.Transaction{
		{Date: "2023-12-05", Description: "Lunch", Amount: 12.50, Type: "withdrawal"},
		{Date: "2023-12-05", Description: "Dinner", Amount: 25.00, Type: "withdrawal"},
	}

	err := client.StoreTransactions(txs)
	if err != nil {
		t.Fatalf("StoreTransactions failed: %v", err)
	}
}

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
