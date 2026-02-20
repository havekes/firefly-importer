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

Use Go 1.22's enhanced `net/http.ServeMux` to handle routing.

1. **`GET /` (IndexHandler):** Parses and executes the `index.html` template, rendering the upload form.
2. **`POST /upload` (UploadHandler):**
   * Parses the multipart form (`r.ParseMultipartForm`).
   * Routes the file to the `parser` package based on its MIME type or extension.
   * Runs the parsed transactions through the deduplication logic.
   * Concurrently or sequentially pushes new transactions via the `firefly` package.
   * Passes the final slice of results to the `index.html` template to render the table.

---

## Phase 5: The User Interface (HTML Templates + Tailwind)

Use Go's `html/template` package to inject data into a single-page UI. We will use the Tailwind CSS Play CDN for rapid prototyping, utilizing utility classes for a "headless" component feel.

* **Template Setup:** Create a `templates/index.html` file. Include `<script src="https://cdn.tailwindcss.com"></script>` in the head.
* **Upload Section:** * A semantic form (`<form enctype="multipart/form-data" method="POST" action="/upload">`).
  * Style the file input and submit button using Tailwind utilities (e.g., `file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100`).
* **Results Section:** * Use Go template tags (`{{ if .Results }}`) to conditionally display a table.
  * Iterate over results (`{{ range .Results }}`).
  * Apply Tailwind classes for row styling based on status: `bg-green-50` for Added, `bg-gray-50` for Skipped, `bg-red-50` for Error.