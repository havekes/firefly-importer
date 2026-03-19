package parser

import (
	"firefly-importer/models"
	"strings"
	"testing"
)

func TestParseCSV(t *testing.T) {
	csvData := `Date,Description,Amount,Type
2023-10-01,Groceries,45.50,withdrawal
2023-10-02,Salary,1500.00,deposit
invalid,bad_row,amount,error`

	r := strings.NewReader(csvData)

	txs, err := ParseCSV(r)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(txs) != 3 {
		t.Errorf("Expected 3 transactions (including 1 error), got %d", len(txs))
	}

	// Verify first transaction
	if txs[0].Date != "2023-10-01" {
		t.Errorf("Expected Date 2023-10-01, got %s", txs[0].Date)
	}
	if txs[0].Description != "Groceries" {
		t.Errorf("Expected Description Groceries, got %s", txs[0].Description)
	}
	if txs[0].Amount != 45.50 {
		t.Errorf("Expected Amount 45.50, got %f", txs[0].Amount)
	}
	if txs[0].Type != "withdrawal" {
		t.Errorf("Expected Type withdrawal, got %s", txs[0].Type)
	}
	if txs[0].Status != models.StatusPending {
		t.Errorf("Expected Status Pending, got %s", txs[0].Status)
	}

	// Verify second transaction
	if txs[1].Date != "2023-10-02" {
		t.Errorf("Expected Date 2023-10-02, got %s", txs[1].Date)
	}
	if txs[1].Amount != 1500.00 {
		t.Errorf("Expected Amount 1500.00, got %f", txs[1].Amount)
	}
	if txs[1].Type != "deposit" {
		t.Errorf("Expected Type deposit, got %s", txs[1].Type)
	}

	// Verify third transaction (the errored one)
	if txs[2].Status != models.StatusError {
		t.Errorf("Expected Status Error, got %s", txs[2].Status)
	}
	if !strings.Contains(txs[2].Description, "Invalid date format") {
		t.Errorf("Expected error message about invalid date, got %s", txs[2].Description)
	}
}
