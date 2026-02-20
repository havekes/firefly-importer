package parser

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"firefly-importer/models"
)

// ParseCSV reads a CSV from the provided io.Reader and maps it to a slice of models.Transaction
// Assumes headers: Date, Description, Amount, Type
func ParseCSV(r io.Reader) ([]models.Transaction, error) {
	csvReader := csv.NewReader(r)

	// Read header row
	_, err := csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("csv file is empty")
		}
		return nil, err
	}

	var transactions []models.Transaction

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if len(record) < 4 {
			continue // Skip incomplete rows
		}

		dateStr := strings.TrimSpace(record[0])

		// Attempt simple YYYY-MM-DD validation
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			// Skip or handle differently, for now just append what we have
		}

		description := strings.TrimSpace(record[1])

		amountStr := strings.TrimSpace(record[2])
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			continue // Skip rows with invalid amounts
		}
		if amount < 0 {
			amount = -amount // Ensure absolute value
		}

		txType := strings.ToLower(strings.TrimSpace(record[3]))

		transactions = append(transactions, models.Transaction{
			Date:        dateStr,
			Description: description,
			Amount:      amount,
			Type:        txType,
			Status:      models.StatusPending,
		})
	}

	return transactions, nil
}
