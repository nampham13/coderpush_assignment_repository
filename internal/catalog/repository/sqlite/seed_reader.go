package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"example.com/supplier-catalog/internal/catalog/domain"
)

func ReadMockProducts(path string) ([]domain.Product, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve mock database path: %w", err)
	}
	source, err := sql.Open("sqlite", "file:"+absolutePath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open mock database: %w", err)
	}
	defer source.Close()
	products, err := readSeedProducts(source)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func readSeedProducts(source *sql.DB) ([]domain.Product, error) {
	rows, err := source.Query(`SELECT product_id, name, description, price_cents, revision, withdrawn
		FROM products ORDER BY product_id LIMIT 501`)
	if err != nil {
		return nil, fmt.Errorf("read products from mock database: %w", err)
	}
	defer rows.Close()
	return scanSeedProducts(rows)
}

func scanSeedProducts(rows *sql.Rows) ([]domain.Product, error) {
	products := make([]domain.Product, 0, 100)
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description,
			&product.PriceCents, &product.Revision, &product.Withdrawn); err != nil {
			return nil, fmt.Errorf("decode mock product: %w", err)
		}
		if len(products) == 500 {
			return nil, errors.New("mock database cannot exceed 500 products")
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mock products: %w", err)
	}
	return products, nil
}
