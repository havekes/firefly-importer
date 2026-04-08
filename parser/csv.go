package parser

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"firefly-importer/models"
)

// ParseCSV reads a CSV from the provided io.Reader and maps it to a slice of models.Transaction
// Expected headers (case-insensitive): Date, Description, Amount, Type
func ParseCSV(r io.Reader) ([]models.Transaction, error) {
	csvReader := csv.NewReader(r)

	// Read header row
	header, err := csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("csv file is empty")
		}
		return nil, err
	}

	// Basic mapping based on column names (case-insensitive)
	colMap := make(map[string]int)
	for i, colName := range header {
		colMap[strings.ToLower(strings.TrimSpace(colName))] = i
	}

	// Validate required columns
	required := []string{"date", "description", "amount", "type"}
	for _, req := range required {
		if _, ok := colMap[req]; !ok {
			return nil, fmt.Errorf("missing required column: %s", req)
		}
	}

	var transactions []models.Transaction

	// Helper to safely get column value by name
	getCol := func(record []string, name string) string {
		idx, ok := colMap[name]
		if !ok || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		dateStr := getCol(record, "date")
		if dateStr == "" {
			continue
		}

		// Attempt simple YYYY-MM-DD validation
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			continue // Skip rows with invalid date formats
		}

		description := getCol(record, "description")
		if description == "" {
			continue
		}

		amountStr := getCol(record, "amount")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			continue // Skip rows with invalid amounts
		}
		if amount < 0 {
			amount = -amount // Ensure absolute value
		}

		txType := strings.ToLower(getCol(record, "type"))
		if txType == "" {
			continue
		}

		transactions = append(transactions, models.Transaction{
			Date:                dateStr,
			Description:         description,
			OriginalDescription: description,
			Amount:              amount,
			Type:                txType,
			Status:              models.StatusPending,
		})
	}

	return transactions, nil
}
