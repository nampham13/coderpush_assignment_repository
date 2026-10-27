package sqlite

import (
	"database/sql"
	"fmt"
	"time"
)

func migrateCatalog(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations(
		version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	var version int
	if err := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version == 1 {
		return nil
	}
	if version > 1 {
		return fmt.Errorf("database schema version %d is newer than supported version 1", version)
	}
	return applyInitialMigration(db)
}

func applyInitialMigration(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin schema migration: %w", err)
	}
	for _, statement := range initialSchema {
		if _, err := tx.Exec(statement); err != nil {
			return rollback(tx, "create catalog schema", err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(1, ?)`, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return rollback(tx, "record schema migration", err)
	}
	return tx.Commit()
}

var initialSchema = []string{
	`CREATE TABLE products (
		product_id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
		price_cents INTEGER NOT NULL CHECK(price_cents >= 0),
		revision INTEGER NOT NULL CHECK(revision > 0), withdrawn INTEGER NOT NULL DEFAULT 0)`,
	`CREATE TABLE updates (id TEXT PRIMARY KEY, status TEXT NOT NULL, body_json TEXT NOT NULL)`,
	`CREATE INDEX products_status_revision ON products(withdrawn, revision)`,
	`CREATE INDEX updates_status ON updates(status)`,
}
