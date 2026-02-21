package models

// Budget represents a Firefly III Budget
type Budget struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BudgetResponse is the wrapper for the Firefly API response
type BudgetResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Name string `json:"name"`
		} `json:"attributes"`
	} `json:"data"`
}

// Category represents a Firefly III Category
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CategoryResponse is the wrapper for the Firefly API response
type CategoryResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Name string `json:"name"`
		} `json:"attributes"`
	} `json:"data"`
}
