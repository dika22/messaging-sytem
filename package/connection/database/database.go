package database

import (
	"database/sql"
	"fmt"
)

type DB struct {
	*sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) CreateTenantPartition(tenantID string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS messages_tenant_%s 
		PARTITION OF messages 
		FOR VALUES IN ('%s')
	`, tenantID, tenantID)

	_, err := db.Exec(query)
	return err
}

func (db *DB) DropTenantPartition(tenantID string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS messages_tenant_%s", tenantID)
	_, err := db.Exec(query)
	return err
}