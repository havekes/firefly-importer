package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"firefly-importer/config"
	"firefly-importer/dedupe"
	"firefly-importer/firefly"
	"firefly-importer/models"
	"firefly-importer/parser"
)

type AppHandler struct {
	Client *firefly.Client
	Config *config.Config
}

func NewAppHandler(client *firefly.Client, cfg *config.Config) *AppHandler {
	return &AppHandler{
		Client: client,
		Config: cfg,
	}
}

// IndexHandler handles GET /
func (h *AppHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.Client.GetAccounts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch accounts: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Welcome to the Firefly III Statement Importer!",
		"accounts": accounts,
	})
}

// UploadHandler handles POST /upload
func (h *AppHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	accountIDStr := r.FormValue("account_id")
	if accountIDStr == "" {
		http.Error(w, "account_id is required", http.StatusBadRequest)
		return
	}
	if _, err := strconv.Atoi(accountIDStr); err != nil {
		http.Error(w, "account_id must be a valid numeric ID", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))

	var parsedTransactions []models.Transaction
	var parseErr error

	switch ext {
	case ".csv":
		parsedTransactions, parseErr = parser.ParseCSV(file)
	case ".png", ".jpg", ".jpeg":
		parsedTransactions, parseErr = parser.ParseImage(file, h.Config.VisionAPIURL, h.Config.VisionAPIKey, h.Config.VisionModel)
	default:
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	if parseErr != nil {
		http.Error(w, fmt.Sprintf("Failed to parse file: %v", parseErr), http.StatusInternalServerError)
		return
	}

	// Fetch existing transactions for deduplication
	existingTransactions, err := h.Client.GetRecentTransactions(30)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch recent transactions: %v", err), http.StatusInternalServerError)
		return
	}

	// Run deduplication filter
	results := dedupe.Filter(parsedTransactions, existingTransactions)

	// Prepare results (set Source/Destination ID without saving)
	for i, tx := range results {
		if tx.Status == models.StatusAdded {
			if strings.ToLower(tx.Type) == "withdrawal" {
				tx.SourceID = accountIDStr
			} else {
				tx.DestinationID = accountIDStr
			}
			results[i] = tx
		}
	}

	// Return parsed results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"results": results,
	})
}

// SaveRequest represents the payload expected by SaveHandler
type SaveRequest struct {
	Transactions []models.Transaction `json:"transactions"`
}

// SaveHandler handles POST /save
func (h *AppHandler) SaveHandler(w http.ResponseWriter, r *http.Request) {
	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return
	}

	addedCount := 0
	errorCount := 0

	for _, tx := range req.Transactions {
		if tx.Status == models.StatusAdded {
			if err := h.Client.StoreTransaction(tx); err != nil {
				log.Printf("Failed to store transaction: %v", err)
				errorCount++
			} else {
				addedCount++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"added":  addedCount,
		"errors": errorCount,
	})
}
