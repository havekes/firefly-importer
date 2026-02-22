package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
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

type PageData struct {
	Accounts    []models.Account
	Budgets     []models.Budget
	Categories  []models.Category
	Results     []models.Transaction
	ResultsJSON template.JS // safe JSON for inline <script>
	Error       string
}

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

// renderPage parses and executes the index.html template with the given data.
func renderPage(w http.ResponseWriter, data PageData) {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		// Headers already sent; log only.
		log.Printf("template execute error: %v", err)
	}
}

// renderError logs the error and renders the index page with an error banner.
// It fetches accounts so the upload form remains functional.
func (h *AppHandler) renderError(w http.ResponseWriter, statusCode int, msg string, err error) {
	if err != nil {
		log.Printf("error: %s: %v", msg, err)
	} else {
		log.Printf("error: %s", msg)
	}

	var errMsg string
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", msg, err)
	} else {
		errMsg = msg
	}

	w.WriteHeader(statusCode)

	accounts, _ := h.Client.GetAccounts() // best-effort; ignore error here
	renderPage(w, PageData{
		Accounts: accounts,
		Error:    errMsg,
	})
}

// IndexHandler handles GET /
func (h *AppHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.Client.GetAccounts()
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to fetch accounts", err)
		return
	}

	renderPage(w, PageData{Accounts: accounts})
}

// UploadHandler handles POST /upload
func (h *AppHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		h.renderError(w, http.StatusBadRequest, "Failed to parse form", err)
		return
	}

	accountIDStr := r.FormValue("account_id")
	if accountIDStr == "" {
		h.renderError(w, http.StatusBadRequest, "account_id is required", nil)
		return
	}
	if _, err := strconv.Atoi(accountIDStr); err != nil {
		h.renderError(w, http.StatusBadRequest, "account_id must be a valid numeric ID", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Failed to get file from form", err)
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
		h.renderError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported file type: %q", ext), nil)
		return
	}

	if parseErr != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to parse file", parseErr)
		return
	}

	// Fetch existing transactions for deduplication
	existingTransactions, err := h.Client.GetRecentTransactions(accountIDStr, 30)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to fetch recent transactions", err)
		return
	}

	// Run deduplication filter
	results := dedupe.Filter(parsedTransactions, existingTransactions)

	// Assign source/destination account ID based on transaction type
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

	// Encode results as JSON for the inline <script> block
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to encode results", err)
		return
	}

	// Fetch accounts again for the form dropdown
	accounts, err := h.Client.GetAccounts()
	if err != nil {
		// non-fatal; the upload result is more important
		log.Printf("Failed to re-fetch accounts: %v", err)
	}

	// Fetch budgets and categories for datalists
	budgets, err := h.Client.GetBudgets()
	if err != nil {
		log.Printf("Failed to re-fetch budgets: %v", err)
	}

	categories, err := h.Client.GetCategories()
	if err != nil {
		log.Printf("Failed to re-fetch categories: %v", err)
	}

	renderPage(w, PageData{
		Accounts:    accounts,
		Budgets:     budgets,
		Categories:  categories,
		Results:     results,
		ResultsJSON: template.JS(jsonBytes),
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
		log.Printf("SaveHandler: failed to parse request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	var toAdd []models.Transaction
	for _, tx := range req.Transactions {
		if tx.Status == models.StatusAdded {
			toAdd = append(toAdd, tx)
		}
	}

	addedCount := 0
	errorCount := 0
	var firstErr error

	if len(toAdd) > 0 {
		if err := h.Client.StoreTransactions(toAdd); err != nil {
			log.Printf("SaveHandler: failed to store transactions: %v", err)
			firstErr = err
			errorCount = len(toAdd)
		} else {
			addedCount = len(toAdd)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if errorCount > 0 && addedCount == 0 {
		// All transactions failed — report as an error
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"added":  addedCount,
			"errors": errorCount,
			"error":  fmt.Sprintf("Failed to save transaction(s). Error: %v", firstErr),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"added":  addedCount,
		"errors": errorCount,
	})
}
