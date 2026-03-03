package firefly

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"firefly-importer/models"
)

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
