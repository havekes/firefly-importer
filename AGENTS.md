# Firefly III Statement Importer - Developer Guide

## Project Overview

This project is a stateless Go web application designed to act as a smart middleman for importing financial transactions into Firefly III. It parses CSV files and screenshots (using an OpenAI-compatible Vision API), deduplicates transactions against an existing Firefly instance, and pushes them via the Firefly III API.

The codebase follows Domain-Driven Design principles with a clean separation of concerns.

## Tech Stack

- **Language**: Go 1.26+
- **Containerization**: Docker & Docker Compose
- **Dependencies**: Standard library `net/http` for the web server (no external web framework), `encoding/csv`, `encoding/json`.
- **External APIs**:
    - Firefly III API (REST)
    - OpenAI-compatible Vision API

## Project Structure

- `cmd/firefly-importer/`: Main application entry point (`main.go`).
- `config/`: Configuration loading (environment variables).
- `dedupe/`: Logic for deduplicating transactions using SHA-256 caching.
- `firefly/`: Client for interacting with the Firefly III API.
- `handlers/`: HTTP request handlers and routing logic.
- `models/`: Shared data structures (Domain Models).
- `parser/`: Parsing logic for CSVs and screenshots (using an OpenAI-compatible Vision API).

## Coding Guidelines & Standards

### General Go Best Practices

- **Formatting**: All code must be formatted with `gofmt`.
- **Nomenclature**:
    - Use `CamelCase` for exported identifiers (public).
    - Use `camelCase` for unexported identifiers (private).
    - Variable names should be short but descriptive (e.g., `i` for loop index, `tx` for transaction).
    - Package names should be short, lowercase, and singular (e.g., `model` not `models`, though existing code uses `models`, follow existing patterns where established, but prefer standard Go idioms for new packages). *Note: The current codebase uses `models`, so stick to that for consistency.*
- **Comments**:
    - Exported functions and types **must** have a comment explaining their purpose.
    - Comments should be complete sentences starting with the function/type name.
- **Error Handling**:
    - **Do not ignore errors**. Handle them or return them.
    - Use `fmt.Errorf("context: %w", err)` to wrap errors and provide context.
    - Avoid `panic` unless during initialization where the app cannot proceed.

### Testing

- **Framework**: Use the standard `testing` package.
- **Pattern**: Prefer **Table-Driven Tests** for testing multiple scenarios.
- **Coverage**: Aim for high test coverage, especially for business logic (`dedupe`, `parser`).
- **Parallelism**: Use `t.Parallel()` for independent tests to speed up execution.

### Concurrency

- Use Go channels and goroutines for concurrent tasks (e.g., processing multiple files).
- Ensure thread safety when accessing shared resources (use `sync.Mutex` or `sync.RWMutex`).
- Avoid race conditions; run tests with `go test -race ./...` to verify.

### Dependency Management

- Use Go Modules (`go.mod`, `go.sum`).
- Run `go mod tidy` after adding or removing dependencies.
- Avoid unnecessary external dependencies; prefer the standard library where possible.

## Workflow

### Local Development

1. **Setup Environment**:
   - Copy `.env.example` to `.env`.
   - Configure `FIREFLY_URL`, `FIREFLY_TOKEN`, etc.

2. **Run Application**:
   ```bash
   go run cmd/firefly-importer/main.go
   ```

3. **Run Tests**:
   ```bash
   go test ./...
   ```
   To run with race detection:
   ```bash
   go test -race ./...
   ```

### Docker Development

1. **Build and Run**:
   ```bash
   docker compose up --build -d
   ```

2. **View Logs**:
   ```bash
   docker compose logs -f
   ```

## Contribution Checklist

Before submitting a change, ensure:
1. The code builds (`go build ./...`).
2. All tests pass (`go test ./...`).
3. Code is formatted (`go fmt ./...`).
4. `go mod tidy` has been run.
5. New functionality is covered by tests.
6. Documentation (comments, README if applicable) is updated.
