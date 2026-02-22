# Firefly III Statement Importer

A stateless Go web application to act as a smart middleman for Firefly III. The app parses CSV files and screenshots, deduplicates transactions against your actual Firefly instance, and pushes them seamlessly via the API.

## Project Structure

The project follows domain-driven design principles with separation of concerns:

- `cmd/firefly-importer/` - Main application entry point (`main.go`).
- `config/` - Handles loading environment variables and configuration.
- `dedupe/` - Deduplication logic mapping imported transactions to existing Firefly records via SHA-256 caching.
- `firefly/` - Communication layer and API client for your Firefly III instance.
- `handlers/` - HTTP request handlers and router setup.
- `models/` - Shared data structs used across the application (e.g., `Transaction`, `Account`).
- `parser/` - File parsing logic handling standard CSV datasets as well as Screenshot OCR via a local OpenAI-compatible Vision API.

## Prerequisites

- **Docker** and **Docker Compose** (Recommended)
- Go 1.22+ (if running on the host natively)
- A running [Firefly III](https://www.firefly-iii.org/) instance and a Personal Access Token
- An OpenAI-compatible Vision API for receipt/screenshot parsing

## Getting Started

### 1. Configuration

Before running the application, you need to set up the environment variables:

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
2. Open `.env` and fill in your details:
   - `FIREFLY_URL`: URL to your Firefly III API (e.g., `https://firefly.example.com/api/v1`)
   - `FIREFLY_TOKEN`: Your Firefly III Personal Access Token
   - `VISION_API_URL`: Your local or hosted Vision API endpoint
   - `VISION_API_KEY`: API Key for the Vision API
   - `PORT`: The port your app will run on (Default `8080`)

### 2. Running the Application locally (Docker Compose)

To start the application in a Docker container, run:

```bash
docker compose up --build -d
```

This will build the Go binary using a multi-stage Dockerfile to keep the image slim, and start the server. 
You can access the UI at `http://localhost:8080` (or whichever port you specified in `.env`).

To view application logs in real-time, run:
```bash
docker compose logs -f
```

To stop the web server:
```bash
docker compose down
```

### 3. Running with Go locally (Alternative)

If you prefer to run it natively without Docker:

```bash
# Download dependencies
go mod download

# Run the project
go run cmd/firefly-importer/main.go
```

The application will be bound to the specified port and act identically to the containerized version.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
