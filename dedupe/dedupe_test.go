package dedupe

import (
	"firefly-importer/models"
	"testing"
)

func TestGenerateHash(t *testing.T) {
	tx1 := models.Transaction{
		Date:        "2023-10-27",
		Description: "Grocery Store",
		Amount:      50.25,
	}

	tx2 := models.Transaction{
		Date:        "2023-10-27",
		Description: "Grocery Store",
		Amount:      50.25,
	}

	tx3 := models.Transaction{
		Date:        "2023-10-27",
		Description: "Grocery Store",
		Amount:      50.26, // different amount
	}

	hash1 := GenerateHash(tx1)
	hash2 := GenerateHash(tx2)
	hash3 := GenerateHash(tx3)

	if hash1 != hash2 {
		t.Errorf("Expected hashes to be identical for identical transactions, got %s and %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("Expected hashes to be different for different transactions, got %s", hash1)
	}
}

func TestFilter(t *testing.T) {
	existing := []models.Transaction{
		{Date: "2023-10-01", Description: "Rent", Amount: 1500.00},
		{Date: "2023-10-02", Description: "Internet", Amount: 60.00},
	}

	incoming := []models.Transaction{
		{Date: "2023-10-01", Description: "Rent", Amount: 1500.00},                         // Duplicate
		{Date: "2023-10-02", Description: "Groceries", Amount: 120.50},                     // New
		{Date: "2023-10-03", Description: "Coffee", Amount: 4.50},                          // New
		{Date: "2023-10-04", Description: "Bad Tx", Amount: 0, Status: models.StatusError}, // Existing error
	}

	result := Filter(incoming, existing)

	if len(result) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(result))
	}

	if result[0].Status != models.StatusSkipped {
		t.Errorf("Expected first transaction to be skipped, got %s", result[0].Status)
	}

	if result[1].Status != models.StatusAdded {
		t.Errorf("Expected second transaction to be added, got %s", result[1].Status)
	}

	if result[2].Status != models.StatusAdded {
		t.Errorf("Expected third transaction to be added, got %s", result[2].Status)
	}

	if result[3].Status != models.StatusError {
		t.Errorf("Expected fourth transaction to retain error status, got %s", result[3].Status)
	}
}
