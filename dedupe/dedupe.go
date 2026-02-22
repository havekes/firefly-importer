package dedupe

import (
	"crypto/sha256"
	"firefly-importer/models"
	"fmt"
)

// GenerateHash generates a SHA-256 hash for a transaction based on Date, Description, and Amount
func GenerateHash(tx models.Transaction, description string) string {
	// We format the amount predictably to avoid floating point inconsistencies
	data := fmt.Sprintf("%s|%s|%.2f|%s", tx.Date, description, tx.Amount, tx.Type)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Filter compares incoming transactions against existing ones and updates their status
func Filter(incoming []models.Transaction, existing []models.Transaction) []models.Transaction {
	// Create a map of existing hashes for O(1) lookup
	existingHashes := make(map[string]bool, len(existing))
	for _, tx := range existing {
		hash := GenerateHash(tx, tx.Description)
		existingHashes[hash] = true
	}

	// Filter incoming transactions
	result := make([]models.Transaction, len(incoming))
	for i, tx := range incoming {
		result[i] = tx // copy by value

		// Skip if it already has an error status
		if result[i].Status == models.StatusError {
			continue
		}

		hash := GenerateHash(tx, tx.Description)
		mappedDescriptionHash := GenerateHash(tx, tx.SuggestedDescription)
		if existingHashes[hash] || existingHashes[mappedDescriptionHash] {
			result[i].Status = models.StatusSkipped
		} else {
			result[i].Status = models.StatusAdded
		}
	}

	return result
}
