package firefly

import (
	"firefly-importer/models"
)

// GetBudgets fetches budgets from Firefly III
func (c *Client) GetBudgets() ([]models.Budget, error) {
	resources, err := c.getPaginatedBasicResources("/budgets")
	if err != nil {
		return nil, err
	}

	var budgets []models.Budget
	for _, item := range resources {
		budgets = append(budgets, models.Budget{
			ID:   item.ID,
			Name: item.Attributes.Name,
		})
	}

	return budgets, nil
}
