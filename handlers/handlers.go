package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"firefly-importer/config"
	"firefly-importer/db"
	"firefly-importer/dedupe"
	"firefly-importer/firefly"
	"firefly-importer/models"
	"firefly-importer/parser"

	"github.com/gorilla/csrf"
)

type PageData struct {
	Accounts    []models.Account
	Budgets     []models.Budget
	Categories  []models.Category
	Results     []models.Transaction
	ResultsJSON string // safe JSON for data attribute
	CSRFField   template.HTML
	CSRFToken   string
	Error       string
}

type AppHandler struct {
	Client *firefly.Client
	Config *config.Config
	DB     *sql.DB
}

func NewAppHandler(client *firefly.Client, cfg *config.Config, dbConn *sql.DB) *AppHandler {
	return &AppHandler{
		Client: client,
		Config: cfg,
		DB:     dbConn,
	}
}

// renderPage executes the pre-parsed index.html template with the given data.
func renderPage(w http.ResponseWriter, r *http.Request, data PageData) {
	data.CSRFField = csrf.TemplateField(r)
	data.CSRFToken = csrf.Token(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := Templates.ExecuteTemplate(w, "index.html", data); err != nil {
		// Headers already sent; log only.
		log.Printf("template execute error: %v", err)
	}
}

// SaveResultData holds data for the save_result.html template snippet
type SaveResultData struct {
	Added int
	Error string
}

// renderSaveResult executes the pre-parsed save_result.html template snippet.
func renderSaveResult(w http.ResponseWriter, data SaveResultData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := Templates.ExecuteTemplate(w, "save-result", data); err != nil {
		log.Printf("template execute error: %v", err)
	}
}

// renderError logs the error and renders the index page with an error banner.
// It fetches accounts so the upload form remains functional.
func (h *AppHandler) renderError(w http.ResponseWriter, r *http.Request, statusCode int, msg string, err error) {
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
	renderPage(w, r, PageData{
		Accounts: accounts,
		Error:    errMsg,
	})
}

// IndexHandler handles GET /
func (h *AppHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.Client.GetAccounts()
	if err != nil {
		h.renderError(w, r, http.StatusInternalServerError, "Failed to fetch accounts", err)
		return
	}

	renderPage(w, r, PageData{Accounts: accounts})
}

// UploadHandler handles POST /upload
func (h *AppHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		h.renderError(w, r, http.StatusBadRequest, "Failed to parse form", err)
		return
	}

	accountIDStr := r.FormValue("account_id")
	if accountIDStr == "" {
		h.renderError(w, r, http.StatusBadRequest, "account_id is required", nil)
		return
	}
	if _, err := strconv.Atoi(accountIDStr); err != nil {
		h.renderError(w, r, http.StatusBadRequest, "account_id must be a valid numeric ID", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.renderError(w, r, http.StatusBadRequest, "Failed to get file from form", err)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	fileDate := r.FormValue("file_date")

	var parsedTransactions []models.Transaction
	var parseErr error

	switch ext {
	case ".csv":
		parsedTransactions, parseErr = parser.ParseCSV(file)
	case ".png", ".jpg", ".jpeg":
		parsedTransactions, parseErr = parser.ParseImage(file, fileDate, h.Config.VisionAPIURL, h.Config.VisionAPIKey, h.Config.VisionModel)
	default:
		h.renderError(w, r, http.StatusBadRequest, fmt.Sprintf("Unsupported file type: %q", ext), nil)
		return
	}

	if parseErr != nil {
		h.renderError(w, r, http.StatusInternalServerError, "Failed to parse file", parseErr)
		return
	}

	// Fetch transaction name mappings and apply them
	mappings, err := db.GetMappings(h.DB)
	if err != nil {
		log.Printf("Failed to fetch name mappings (ignoring): %v", err)
	} else if len(mappings) > 0 {
		for i, tx := range parsedTransactions {
			if m, ok := mappings[tx.OriginalDescription]; ok {
				parsedTransactions[i].SuggestedDescription = m.NewName
				parsedTransactions[i].SuggestedBudget = m.BudgetName
				parsedTransactions[i].SuggestedCategory = m.CategoryName
			}
		}
	}

	// Fetch existing transactions for deduplication
	existingTransactions, err := h.Client.GetRecentTransactions(accountIDStr, 30)
	if err != nil {
		h.renderError(w, r, http.StatusInternalServerError, "Failed to fetch recent transactions", err)
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
		h.renderError(w, r, http.StatusInternalServerError, "Failed to encode results", err)
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

	renderPage(w, r, PageData{
		Accounts:    accounts,
		Budgets:     budgets,
		Categories:  categories,
		Results:     results,
		ResultsJSON: string(jsonBytes),
	})
}

// SaveRequest represents the payload expected by SaveHandler
type SaveRequest struct {
	Transactions []models.Transaction `json:"transactions"`
}

// SaveHandler handles POST /save
func (h *AppHandler) SaveHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("SaveHandler: failed to parse form: %v", err)
		renderSaveResult(w, SaveResultData{Error: "Failed to parse form submission"})
		return
	}

	payload := r.FormValue("payload")
	if payload == "" {
		renderSaveResult(w, SaveResultData{Error: "No payload provided"})
		return
	}

	var req SaveRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		log.Printf("SaveHandler: failed to parse payload JSON: %v", err)
		renderSaveResult(w, SaveResultData{Error: fmt.Sprintf("Failed to parse request payload: %v", err)})
		return
	}

	addedCount := 0
	errorCount := 0
	var firstErr error

	for _, tx := range req.Transactions {
		if tx.Status == models.StatusAdded {
			if err := h.Client.StoreTransaction(tx); err != nil {
				log.Printf("SaveHandler: failed to store transaction %q: %v", tx.Description, err)
				if firstErr == nil {
					firstErr = err
				}
				errorCount++
			} else {
				addedCount++
				// If the description was edited mapping to a new name or budget/category were added, save the mapping
				if tx.OriginalDescription != "" && (tx.OriginalDescription != tx.Description || tx.BudgetName != "" || tx.CategoryName != "") {
					if err := db.SaveMapping(h.DB, tx.OriginalDescription, tx.Description, tx.BudgetName, tx.CategoryName); err != nil {
						log.Printf("Failed to save mapping for %q -> %q, Budget: %q, Category: %q: %v", tx.OriginalDescription, tx.Description, tx.BudgetName, tx.CategoryName, err)
					}
				}
			}
		}
	}

	if errorCount > 0 && addedCount == 0 {
		// All transactions failed — report as an error
		renderSaveResult(w, SaveResultData{
			Error: fmt.Sprintf("All %d transaction(s) failed to save. First error: %v", errorCount, firstErr),
		})
		return
	}

	if errorCount > 0 {
		renderSaveResult(w, SaveResultData{
			Added: addedCount,
			Error: fmt.Sprintf("Saved %d, but %d failed. First error: %v", addedCount, errorCount, firstErr),
		})
		return
	}

	// Success
	renderSaveResult(w, SaveResultData{Added: addedCount})
}
