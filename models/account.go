package models

// Account represents a Firefly III account.
type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// AccountResponse wrapper for the Firefly API JSON response
type AccountResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"attributes"`
	} `json:"data"`
}
