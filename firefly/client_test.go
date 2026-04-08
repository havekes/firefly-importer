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

func TestGetResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		endpoint     string
		mockResponse string
		callFunc     func(*Client) (any, error)
		expectedLen  int
		verifyFirst  func(t *testing.T, item any)
	}{
		{
			name:     "GetBudgets",
			endpoint: "/budgets",
			mockResponse: `{
				"data": [{"id": "10", "attributes": {"name": "Groceries"}}],
				"meta": {"pagination": {"total_pages": 1, "current_page": 1}}
			}`,
			callFunc: func(c *Client) (any, error) { return c.GetBudgets() },
			expectedLen: 1,
			verifyFirst: func(t *testing.T, item any) {
				budget := item.(models.Budget)
				if budget.ID != "10" {
					t.Errorf("Expected ID 10, got %s", budget.ID)
				}
				if budget.Name != "Groceries" {
					t.Errorf("Expected Name Groceries, got %s", budget.Name)
				}
			},
		},
		{
			name:     "GetCategories",
			endpoint: "/categories",
			mockResponse: `{
				"data": [{"id": "20", "attributes": {"name": "Entertainment"}}],
				"meta": {"pagination": {"total_pages": 1, "current_page": 1}}
			}`,
			callFunc: func(c *Client) (any, error) { return c.GetCategories() },
			expectedLen: 1,
			verifyFirst: func(t *testing.T, item any) {
				category := item.(models.Category)
				if category.ID != "20" {
					t.Errorf("Expected ID 20, got %s", category.ID)
				}
				if category.Name != "Entertainment" {
					t.Errorf("Expected Name Entertainment, got %s", category.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.endpoint {
					t.Errorf("Expected path %s, got %s", tt.endpoint, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.Write([]byte(tt.mockResponse))
			}))
			defer mockServer.Close()

			client := NewClient(mockServer.URL, "test-token")
			result, err := tt.callFunc(client)

			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}

			// Using reflection to check length as result is any (slice)
			switch items := result.(type) {
			case []models.Budget:
				if len(items) != tt.expectedLen {
					t.Fatalf("Expected %d budgets, got %d", tt.expectedLen, len(items))
				}
				if tt.expectedLen > 0 {
					tt.verifyFirst(t, items[0])
				}
			case []models.Category:
				if len(items) != tt.expectedLen {
					t.Fatalf("Expected %d categories, got %d", tt.expectedLen, len(items))
				}
				if tt.expectedLen > 0 {
					tt.verifyFirst(t, items[0])
				}
			default:
				t.Fatalf("Unknown result type: %T", result)
			}
		})
	}
}
