package models

// Budget represents a Firefly III Budget
type Budget struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Category represents a Firefly III Category
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
