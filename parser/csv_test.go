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
		verifyFirst   *models.Transaction
	}{
		{
			name: "Valid CSV",
			csvData: `Date,Description,Amount,Type
2023-10-01,Groceries,45.50,withdrawal
2023-10-02,Salary,1500.00,deposit
invalid,bad_row,amount,error`,
			expectedLen: 2,
			expectError: false,
			verifyFirst: &models.Transaction{
				Date:        "2023-10-01",
				Description: "Groceries",
				Amount:      45.50,
				Type:        "withdrawal",
				Status:      models.StatusPending,
			},
		},
		{
			name: "Different Column Order",
			csvData: `Amount,Type,Description,Date
45.50,withdrawal,Groceries,2023-10-01
1500.00,deposit,Salary,2023-10-02`,
			expectedLen: 2,
			expectError: false,
			verifyFirst: &models.Transaction{
				Date:        "2023-10-01",
				Description: "Groceries",
				Amount:      45.50,
				Type:        "withdrawal",
				Status:      models.StatusPending,
			},
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

			if tt.verifyFirst != nil {
				if len(txs) == 0 {
					t.Fatalf("Expected at least 1 transaction to verify, got 0")
				}
				got := txs[0]
				want := tt.verifyFirst
				if got.Date != want.Date {
					t.Errorf("Expected Date %s, got %s", want.Date, got.Date)
				}
				if got.Description != want.Description {
					t.Errorf("Expected Description %s, got %s", want.Description, got.Description)
				}
				if got.Amount != want.Amount {
					t.Errorf("Expected Amount %f, got %f", want.Amount, got.Amount)
				}
				if got.Type != want.Type {
					t.Errorf("Expected Type %s, got %s", want.Type, got.Type)
				}
				if got.Status != want.Status {
					t.Errorf("Expected Status %s, got %s", want.Status, got.Status)
				}
			}
		})
	}
}
