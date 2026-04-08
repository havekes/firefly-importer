package parser

import (
	"firefly-importer/models"
	"strings"
	"testing"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name          string
		csvData       string
		expectedLen   int
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid CSV",
			csvData: `Date,Description,Amount,Type
2023-10-01,Groceries,45.50,withdrawal
2023-10-02,Salary,1500.00,deposit
invalid,bad_row,amount,error`,
			expectedLen: 2,
			expectError: false,
		},
		{
			name: "Different Column Order",
			csvData: `Amount,Type,Description,Date
45.50,withdrawal,Groceries,2023-10-01
1500.00,deposit,Salary,2023-10-02`,
			expectedLen: 2,
			expectError: false,
		},
		{
			name: "Missing Amount Column",
			csvData: `Date,Description,Type
2023-10-01,Groceries,withdrawal`,
			expectError:   true,
			errorContains: "missing required column: amount",
		},
		{
			name: "Empty CSV",
			csvData: ``,
			expectError:   true,
			errorContains: "csv file is empty",
		},
		{
			name: "Rows with missing data",
			csvData: `Date,Description,Amount,Type
2023-10-01,Groceries,45.50,withdrawal
,Missing Date,10.00,withdrawal
2023-10-02,,20.00,withdrawal
2023-10-03,Missing Amount,,withdrawal
2023-10-04,Missing Type,30.00,`,
			expectedLen: 1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.csvData)
			txs, err := ParseCSV(r)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorContains)) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseCSV failed: %v", err)
			}

			if len(txs) != tt.expectedLen {
				t.Errorf("Expected %d transactions, got %d", tt.expectedLen, len(txs))
			}

			if len(txs) > 0 && tt.name == "Valid CSV" {
				// Verify first transaction of Valid CSV
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
			}
		})
	}
}
