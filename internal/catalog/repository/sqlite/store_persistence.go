package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"example.com/supplier-catalog/internal/catalog/domain"
)

func (r *Repository) Save(next domain.State) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin catalog transaction: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM products`); err != nil {
		return rollback(tx, "clear products", err)
	}
	for _, product := range next.Catalog {
		if err := insertProduct(tx, product); err != nil {
			return rollback(tx, "save product", err)
		}
	}
	if _, err := tx.Exec(`DELETE FROM updates`); err != nil {
		return rollback(tx, "clear updates", err)
	}
	for _, update := range next.Updates {
		body, err := json.Marshal(update)
		if err != nil {
			return rollback(tx, "encode update", err)
		}
		if _, err := tx.Exec(`INSERT INTO updates(id, status, body_json) VALUES(?, ?, ?)`, update.ID, update.Status, body); err != nil {
			return rollback(tx, "save update", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit catalog transaction: %w", err)
	}
	return nil
}

func rollback(tx *sql.Tx, action string, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
		return fmt.Errorf("%s: %w (rollback failed: %v)", action, err, rollbackErr)
	}
	return fmt.Errorf("%s: %w", action, err)
}

func insertProduct(tx *sql.Tx, product domain.Product) error {
	_, err := tx.Exec(`INSERT INTO products(product_id, name, description, price_cents, revision, withdrawn)
		VALUES(?, ?, ?, ?, ?, ?)`, product.ID, product.Name, product.Description,
		product.PriceCents, product.Revision, product.Withdrawn)
	return err
}
