package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var logQueries bool

// EnableQueryLogging toggles logging of database queries.
func EnableQueryLogging(enable bool) {
	logQueries = enable
}

// InitDB connects to the PostgreSQL database and creates the necessary tables.
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database and verified schema.")
	return db, nil
}

// Mapping represents a saved mapping for description, budget, and category.
type Mapping struct {
	OriginalName string
	NewName      string
	BudgetName   string
	CategoryName string
}

//go:embed schema.sql
var schemaSQL string

func createSchema(db *sql.DB) error {
	if logQueries {
		log.Printf("[DB DEBUG] Executing schema SQL")
	}
	_, err := db.Exec(schemaSQL)
	return err
}

// SaveMapping inserts or updates a name mapping from original to new name, budget, and category.
func SaveMapping(db *sql.DB, original, newDesc, budget, category string) error {
	if db == nil {
		return nil
	}
	query := `
	INSERT INTO name_mappings (original_name, new_name, budget_name, category_name, updated_at)
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	ON CONFLICT (original_name) 
	DO UPDATE SET new_name = EXCLUDED.new_name, budget_name = EXCLUDED.budget_name, category_name = EXCLUDED.category_name, updated_at = EXCLUDED.updated_at;
	`
	if logQueries {
		log.Printf("[DB DEBUG] Executing: %s args: [%s, %s, %s, %s]", query, original, newDesc, budget, category)
	}
	_, err := db.Exec(query, original, newDesc, budget, category)
	if err != nil {
		return fmt.Errorf("failed to upsert name mapping: %w", err)
	}
	return nil
}

// GetMappings retrieves all name mappings from the database.
func GetMappings(db *sql.DB) (map[string]Mapping, error) {
	if db == nil {
		return nil, nil
	}
	mappings := make(map[string]Mapping)

	query := `SELECT original_name, new_name, COALESCE(budget_name, ''), COALESCE(category_name, '') FROM name_mappings;`
	if logQueries {
		log.Printf("[DB DEBUG] Executing: %s", query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query name mappings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Mapping
		if err := rows.Scan(&m.OriginalName, &m.NewName, &m.BudgetName, &m.CategoryName); err != nil {
			return nil, fmt.Errorf("failed to scan name mapping row: %w", err)
		}
		mappings[m.OriginalName] = m
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating name mapping rows: %w", err)
	}

	return mappings, nil
}
