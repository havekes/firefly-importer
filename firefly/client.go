package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
				Date            string `json:"date"`
				Description     string `json:"description"`
				Amount          string `json:"amount"` // Note: Firefly amount is often a string
				Type            string `json:"type"`
				SourceName      string `json:"source_name"`
				DestinationName string `json:"destination_name"`
			} `json:"transactions"`
		} `json:"attributes"`
	} `json:"data"`
}

// GetRecentTransactions fetches recent transactions for deduplication purposes
func (c *Client) GetRecentTransactions(accountID string, daysOffset int) ([]models.Transaction, error) {
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -daysOffset).Format("2006-01-02")
	req, err := http.NewRequest("GET", c.BaseURL+"/accounts/"+accountID+"/transactions?start="+startDate+"&end="+endDate, nil)
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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code %d and failed to read response body: %w", resp.StatusCode, err)
		}
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
			amount, amountErr := strconv.ParseFloat(tx.Amount, 64)
			if amountErr != nil {
				// Consider logging this error
				continue // Skip transaction with unparseable amount
			}

			parsedDate, dateErr := time.Parse(time.RFC3339, tx.Date)
			if dateErr != nil {
				// Consider logging this error
				continue // Skip transaction with unparseable date
			}

			transactions = append(transactions, models.Transaction{
				Date:            parsedDate.Format("2006-01-02"),
				Description:     tx.Description,
				Amount:          amount,
				Type:            tx.Type,
				SourceName:      tx.SourceName,
				DestinationName: tx.DestinationName,
				Status:          models.StatusAdded, // existing transactions are "added"
			})
		}
	}

	return transactions, nil
}

// GetAccounts fetches asset accounts from Firefly III
func (c *Client) GetAccounts() ([]models.Account, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/accounts?type=asset", nil)
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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code %d and failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fireflyResp models.AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&fireflyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var accounts []models.Account
	for _, item := range fireflyResp.Data {
		accounts = append(accounts, models.Account{
			ID:   item.ID,
			Name: item.Attributes.Name,
			Type: item.Attributes.Type,
		})
	}

	return accounts, nil
}

// GetBudgets fetches budgets from Firefly III
func (c *Client) GetBudgets() ([]models.Budget, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/budgets", nil)
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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code %d and failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fireflyResp models.BudgetResponse
	if err := json.NewDecoder(resp.Body).Decode(&fireflyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var budgets []models.Budget
	for _, item := range fireflyResp.Data {
		budgets = append(budgets, models.Budget{
			ID:   item.ID,
			Name: item.Attributes.Name,
		})
	}

	return budgets, nil
}

// GetCategories fetches categories from Firefly III
func (c *Client) GetCategories() ([]models.Category, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/categories", nil)
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
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status code %d and failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fireflyResp models.CategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&fireflyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var categories []models.Category
	for _, item := range fireflyResp.Data {
		categories = append(categories, models.Category{
			ID:   item.ID,
			Name: item.Attributes.Name,
		})
	}

	return categories, nil
}

// fireflyStoreTransactionRequest represents the payload to create a new transaction
type fireflyStoreTransactionRequest struct {
	Transactions []storeTx `json:"transactions"`
}

type storeTx struct {
	Date            string `json:"date"` // YYYY-MM-DD
	Description     string `json:"description"`
	Amount          string `json:"amount"`
	Type            string `json:"type"`
	SourceName      string `json:"source_name,omitempty"`
	SourceID        string `json:"source_id,omitempty"`
	DestinationName string `json:"destination_name,omitempty"`
	DestinationID   string `json:"destination_id,omitempty"`
	BudgetName      string `json:"budget_name,omitempty"`
	CategoryName    string `json:"category_name,omitempty"`
}

// StoreTransaction posts a single transaction to Firefly III
func (c *Client) StoreTransaction(tx models.Transaction) error {
	payload := fireflyStoreTransactionRequest{
		Transactions: []storeTx{
			{
				Date:            tx.Date,
				Description:     tx.Description,
				Amount:          fmt.Sprintf("%.2f", tx.Amount),
				Type:            tx.Type,
				SourceName:      tx.SourceName,
				SourceID:        tx.SourceID,
				DestinationName: tx.DestinationName,
				DestinationID:   tx.DestinationID,
				BudgetName:      tx.BudgetName,
				CategoryName:    tx.CategoryName,
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
		respBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code %d and failed to read response body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBodyBytes))
	}

	return nil
}
