package sqlite

import (
	"encoding/json"
	"fmt"

	"example.com/supplier-catalog/internal/catalog/domain"
)

func (r *Repository) Load() (domain.State, error) {
	state := domain.State{Catalog: map[string]domain.Product{}, Updates: map[string]*domain.Update{}}
	if err := r.loadProducts(state.Catalog); err != nil {
		return domain.State{}, err
	}
	if err := r.loadUpdates(state.Updates); err != nil {
		return domain.State{}, err
	}
	return state, nil
}

func (r *Repository) loadProducts(products map[string]domain.Product) error {
	rows, err := r.db.Query(`SELECT product_id, name, description, price_cents, revision, withdrawn FROM products`)
	if err != nil {
		return fmt.Errorf("read products: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description,
			&product.PriceCents, &product.Revision, &product.Withdrawn); err != nil {
			return fmt.Errorf("decode product row: %w", err)
		}
		products[product.ID] = product
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate products: %w", err)
	}
	return nil
}

func (r *Repository) loadUpdates(updates map[string]*domain.Update) error {
	rows, err := r.db.Query(`SELECT body_json FROM updates`)
	if err != nil {
		return fmt.Errorf("read updates: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var body []byte
		var update domain.Update
		if err := rows.Scan(&body); err != nil {
			return fmt.Errorf("read update row: %w", err)
		}
		if err := json.Unmarshal(body, &update); err != nil {
			return fmt.Errorf("decode update row: %w", err)
		}
		updates[update.ID] = &update
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate updates: %w", err)
	}
	return nil
}
