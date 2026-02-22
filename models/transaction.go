package models

// TransactionStatus represents the state of a transaction during processing
type TransactionStatus string

const (
	StatusPending TransactionStatus = "Pending"
	StatusAdded   TransactionStatus = "Added"
	StatusSkipped TransactionStatus = "Skipped (Duplicate)"
	StatusError   TransactionStatus = "Error"
)

// Transaction represents a single financial transaction
type Transaction struct {
	Date                 string            `json:"date"` // Format: YYYY-MM-DD
	Description          string            `json:"description"`
	OriginalDescription  string            `json:"original_description,omitempty"`
	SuggestedDescription string            `json:"suggested_description,omitempty"`
	Amount               float64           `json:"amount"` // Absolute value
	Type                 string            `json:"type"`   // "withdrawal" or "deposit"
	SourceName           string            `json:"source_name,omitempty"`
	SourceID             string            `json:"source_id,omitempty"`
	DestinationName      string            `json:"destination_name,omitempty"`
	DestinationID        string            `json:"destination_id,omitempty"`
	BudgetName           string            `json:"budget_name,omitempty"`
	SuggestedBudget      string            `json:"suggested_budget,omitempty"`
	CategoryName         string            `json:"category_name,omitempty"`
	SuggestedCategory    string            `json:"suggested_category,omitempty"`
	Status               TransactionStatus `json:"status,omitempty"`
}
