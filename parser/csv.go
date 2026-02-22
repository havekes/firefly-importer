package parser

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
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
			log.Printf("CSV parse error: incomplete row: %v", record)
			transactions = append(transactions, models.Transaction{
				Description: fmt.Sprintf("Incomplete row: %v", record),
				Status:      models.StatusError,
			})
			continue
		}

		dateStr := strings.TrimSpace(record[0])

		// Attempt simple YYYY-MM-DD validation
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			log.Printf("CSV parse error: invalid date format %q: %v", dateStr, err)
			transactions = append(transactions, models.Transaction{
				Date:        dateStr,
				Description: fmt.Sprintf("Invalid date format: %v", err),
				Status:      models.StatusError,
			})
			continue
		}

		description := strings.TrimSpace(record[1])

		amountStr := strings.TrimSpace(record[2])
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Printf("CSV parse error: invalid amount %q: %v", amountStr, err)
			transactions = append(transactions, models.Transaction{
				Date:        dateStr,
				Description: fmt.Sprintf("Invalid amount %q: %v", amountStr, err),
				Status:      models.StatusError,
			})
			continue
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
