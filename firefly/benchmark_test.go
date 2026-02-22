package firefly

import (
	"encoding/json"
	"firefly-importer/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkStoreTransactions(b *testing.B) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqPayload fireflyStoreTransactionRequest
		json.NewDecoder(r.Body).Decode(&reqPayload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, "test-token")

	tx := models.Transaction{
		Date:            "2023-12-05",
		Description:     "Lunch",
		Amount:          12.50,
		Type:            "withdrawal",
		SourceName:      "Wallet",
		DestinationName: "Restaurant",
	}

	b.Run("Single-Sequential-10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 10; j++ {
				_ = client.StoreTransaction(tx)
			}
		}
	})

	b.Run("Batched-10", func(b *testing.B) {
		txs := make([]models.Transaction, 10)
		for j := 0; j < 10; j++ {
			txs[j] = tx
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.StoreTransactions(txs)
		}
	})
}
