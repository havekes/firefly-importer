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
	Date        string            `json:"date"` // Format: YYYY-MM-DD
	Description string            `json:"description"`
	Amount      float64           `json:"amount"` // Absolute value
	Type        string            `json:"type"`   // "withdrawal" or "deposit"
	Status      TransactionStatus `json:"status,omitempty"`
}
