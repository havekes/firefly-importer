package firefly

import (
	"firefly-importer/models"
)

// GetCategories fetches categories from Firefly III
func (c *Client) GetCategories() ([]models.Category, error) {
	resources, err := c.getPaginatedBasicResources("/categories")
	if err != nil {
		return nil, err
	}

	var categories []models.Category
	for _, item := range resources {
		categories = append(categories, models.Category{
			ID:   item.ID,
			Name: item.Attributes.Name,
		})
	}

	return categories, nil
}
