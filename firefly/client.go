package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"firefly-importer/models"
)

// Client handles communication with the Firefly III API
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new Firefly III API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// fireflyTransactionResponse represents the response format for getting transactions
type fireflyTransactionResponse struct {
	Data []struct {
		Attributes struct {
			Transactions []struct {
				Date        string `json:"date"`
				Description string `json:"description"`
				Amount      string `json:"amount"` // Note: Firefly amount is often a string
				Type        string `json:"type"`
			} `json:"transactions"`
		} `json:"attributes"`
	} `json:"data"`
}

// GetRecentTransactions fetches recent transactions for deduplication purposes
func (c *Client) GetRecentTransactions() ([]models.Transaction, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/transactions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fireflyResp fireflyTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&fireflyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var transactions []models.Transaction

	for _, item := range fireflyResp.Data {
		for _, tx := range item.Attributes.Transactions {
			// Convert string amount to float handling
			var amount float64
			fmt.Sscanf(tx.Amount, "%f", &amount)

			transactions = append(transactions, models.Transaction{
				Date:        tx.Date[:10], // Assuming date format like 2023-01-01T00:00:00Z
				Description: tx.Description,
				Amount:      amount,
				Type:        tx.Type,
				Status:      models.StatusAdded, // existing transactions are "added"
			})
		}
	}

	return transactions, nil
}

// fireflyStoreTransactionRequest represents the payload to create a new transaction
type fireflyStoreTransactionRequest struct {
	Transactions []storeTx `json:"transactions"`
}

type storeTx struct {
	Date        string `json:"date"` // YYYY-MM-DD
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Type        string `json:"type"`
	// Additional fields like source/destination accounts would go here in a full app
}

// StoreTransaction posts a single transaction to Firefly III
func (c *Client) StoreTransaction(tx models.Transaction) error {
	payload := fireflyStoreTransactionRequest{
		Transactions: []storeTx{
			{
				Date:        tx.Date,
				Description: tx.Description,
				Amount:      fmt.Sprintf("%.2f", tx.Amount),
				Type:        tx.Type,
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/transactions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBodyBytes))
	}

	return nil
}
