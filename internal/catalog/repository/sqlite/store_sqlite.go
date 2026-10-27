package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func Open(path string) (*Repository, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	db.SetMaxOpenConns(1)
	repository := &Repository{db: db}
	if err := repository.Migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if path != ":memory:" {
		if err := os.Chmod(path, 0o600); err != nil {
			db.Close()
			return nil, fmt.Errorf("secure SQLite database: %w", err)
		}
	}
	return repository, nil
}

func (s *Repository) Migrate() error {
	for _, pragma := range []string{"PRAGMA busy_timeout = 5000", "PRAGMA journal_mode = WAL", "PRAGMA foreign_keys = ON"} {
		if _, err := s.db.Exec(pragma); err != nil {
			return fmt.Errorf("configure SQLite: %w", err)
		}
	}
	return migrateCatalog(s.db)
}

func (r *Repository) Close() error {
	return r.db.Close()
}
