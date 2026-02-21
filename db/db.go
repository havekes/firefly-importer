package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

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

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS name_mappings (
		original_name TEXT PRIMARY KEY,
		new_name TEXT NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(query)
	return err
}

// SaveMapping inserts or updates a name mapping from original to new name.
func SaveMapping(db *sql.DB, original, new string) error {
	if db == nil {
		return nil
	}
	query := `
	INSERT INTO name_mappings (original_name, new_name, updated_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	ON CONFLICT (original_name) 
	DO UPDATE SET new_name = EXCLUDED.new_name, updated_at = EXCLUDED.updated_at;
	`
	_, err := db.Exec(query, original, new)
	if err != nil {
		return fmt.Errorf("failed to upsert name mapping: %w", err)
	}
	return nil
}

// GetMappings retrieves all name mappings from the database.
func GetMappings(db *sql.DB) (map[string]string, error) {
	if db == nil {
		return nil, nil
	}
	mappings := make(map[string]string)

	query := `SELECT original_name, new_name FROM name_mappings;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query name mappings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var original, new string
		if err := rows.Scan(&original, &new); err != nil {
			return nil, fmt.Errorf("failed to scan name mapping row: %w", err)
		}
		mappings[original] = new
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating name mapping rows: %w", err)
	}

	return mappings, nil
}
