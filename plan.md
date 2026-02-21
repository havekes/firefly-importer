# Firefly III Statement Importer: Go Implementation Plan

A stateless Go web application to act as a smart middleman for Firefly III. The app will parse CSVs and screenshots, deduplicate transactions, and push them to Firefly via API.



## Phase 1: Project Setup and Configuration

Since this is a stateless prototype, no database is needed. We will rely on Go's standard library for routing (`net/http`) and templating (`html/template`).

1. **Initialize Go Module:**
   * Run `mkdir firefly-importer && cd firefly-importer`.
   * Run `go mod init firefly-importer`.
2. **Environment Variables:**
   * Use a package like `github.com/joho/godotenv` to load a `.env` file during development.
   * Add the necessary keys to your `.env` file:
     * `FIREFLY_URL="https://firefly.havek.es/api/v1"`
     * `FIREFLY_TOKEN=` (Your Personal Access Token from Firefly III)
     * `VISION_API_URL="https://ai.havek.es/api"` (Local OpenAI-compatible endpoint)
     * `VISION_API_KEY=` (To be provided later)
     * `PORT="8080"`

---

## Phase 2: Service Layer Architecture (Structs & Methods)

Instead of classes, we will define Go structs and interfaces to handle the business logic cleanly.

### 1. `parser` Package
This package takes the uploaded `multipart.File` and returns a slice of `Transaction` structs (`Date`, `Description`, `Amount`, `Type`).
* **CSV Handling:** Use Go's standard `encoding/csv` package. Iterate through the records, parse the strings, and map them to your `Transaction` struct.
* **Screenshot Handling:**
  * Read the image into memory and encode it to a Base64 string (`encoding/base64`).
  * Construct a JSON payload matching the OpenAI API spec.
  * Send an HTTP POST request to your local `VISION_API_URL` using `net/http`. 
  * Prompt: *"Extract bank transactions from this image. Return ONLY a JSON array with objects containing: date (YYYY-MM-DD), description (string), amount (float, absolute value), and type (withdrawal or deposit)."*
  * Unmarshal (`encoding/json`) the response text back into a slice of `Transaction` structs.

### 2. `firefly` Package
This package handles communication with your Firefly III instance.
* **`GetRecentTransactions()`:** Makes a GET request to `/api/v1/transactions` with the Authorization header to fetch the recent ledger for deduplication.
* **`StoreTransaction(tx Transaction)`:** Wraps your parsed struct into the Firefly-specific JSON payload (requiring the `transactions` array and account details) and sends a POST request.

---

## Phase 3: Deduplication Logic

1. **Fetch Existing:** Before processing the upload, call `firefly.GetRecentTransactions()`.
2. **Create Hashes:** For both existing transactions and your newly parsed slice, generate a unique hash using `crypto/sha256`.
   * *Example:* `fmt.Sprintf("%x", sha256.Sum256([]byte(date + description + amount)))`
3. **Filter:** Iterate through your parsed `Transaction` slice. Check if the generated hash exists in a map of Firefly hashes (using a `map[string]bool` for quick `O(1)` lookups). Tag the struct's `Status` field as either `Added`, `Skipped (Duplicate)`, or `Error`.

---

## Phase 4: HTTP Handlers (Routing)

Use Go 1.22's enhanced `net/http.ServeMux` to handle routing. You'll need to fetch the user's accounts on the initial page load so they can select the target account for the import.

1. **`GET /` (IndexHandler):** * Calls a new `firefly.GetAccounts()` method to fetch a list of your asset/bank accounts from Firefly III (via `GET /api/v1/accounts?type=asset`).
   * Parses the `index.html` template and passes the retrieved slice of `Account` structs to it.
2. **`POST /upload` (UploadHandler):**
   * Parses the multipart form (`r.ParseMultipartForm`).
   * Extracts the `account_id` selected by the user from the form data (`r.FormValue("account_id")`).
   * Routes the file to the `parser` package based on its MIME type or extension.
   * Runs the parsed transactions through the deduplication logic.
   * Pushes new transactions via `firefly.StoreTransaction(tx, accountID)`, ensuring the selected account is set as the source or destination depending on the transaction type.
   * Passes the final slice of results back to the `index.html` template to render the table.

---

## Phase 5: The User Interface (HTML Templates + Tailwind)

Use Go's `html/template` package to inject data into the single-page UI. We will use the Tailwind CSS CDN for rapid styling.

* **Template Setup:** Create a `templates/index.html` file. Include `<script src="https://cdn.tailwindcss.com"></script>` in the head.
* **Upload Section:** * Create a semantic form: `<form enctype="multipart/form-data" method="POST" action="/upload" class="space-y-4">`.
  * **Account Dropdown:** Add a select menu for the target account. Use Go template tags to iterate over the accounts fetched in Phase 4:
    ```html
    <select name="account_id" class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm">
      {{ range .Accounts }}
        <option value="{{ .ID }}">{{ .Name }}</option>
      {{ end }}
    </select>
    ```
  * **File Input:** Style the file input using Tailwind utilities (e.g., `file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100`).
  * **Submit Button:** A standard styled button (`bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded`).
* **Results Section:** * Use Go template tags (`{{ if .Results }}`) to conditionally display a table.
  * Iterate over results (`{{ range .Results }}`).
  * Apply Tailwind classes for row styling based on status: `bg-green-50` for Added, `bg-yellow-50` for Skipped (Duplicate), `bg-red-50` for Error.

---

## Phase 6: Bulk Selection of Budgets and Categories (Enrichment)

To provide more control over the data being imported, users should be able to assign a budget and a category to each transaction before saving.

### 1. Fetching Budgets and Categories
* Add a `GetBudgets()` method to the `firefly` package that calls `GET /api/v1/budgets` (per the [Budgets API Docs](https://api-docs.firefly-iii.org/#/budgets/listBudget)).
* Add a `GetCategories()` method to the `firefly` package that calls `GET /api/v1/categories` (per the [Categories API Docs](https://api-docs.firefly-iii.org/#/categories/listCategory)).
* Pass the retrieved budgets and categories to the frontend template when rendering the parsed transactions.

### 2. Frontend Autocomplete Inputs
* In the results/review table, add column(s) for Budget and Category on each proposed transaction.
* Implement an **autocomplete field** for the budget and category inputs using the fetched lists.
* This allows the user to bulk select or quickly apply an existing budget/category by typing, ensuring valid selection before the final submission.

### 3. Storing Enriched Transactions
* When the user submits the selected transactions to be saved, capture the assigned budget and category IDs/names from the form payload.
* Update the `StoreTransaction` method (which uses `POST /api/v1/transactions` - [Store Transaction API Docs](https://api-docs.firefly-iii.org/#/transactions/storeTransaction)) to include `budget_id` (or `budget_name`) and `category_id` (or `category_name`) in the request body for each transaction.