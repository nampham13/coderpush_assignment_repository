package service

import (
	"fmt"
	"strings"
)

func (s *Store) SeedProducts(products []Product) (int, error) {
	inserted := 0
	err := s.change(func(next *state) error {
		seen := map[string]bool{}
		for _, product := range products {
			product.ID = strings.TrimSpace(product.ID)
			if err := validateSeedProduct(product); err != nil {
				return err
			}
			if seen[product.ID] {
				return fmt.Errorf("%w: duplicate seed product %q", errInvalid, product.ID)
			}
			seen[product.ID] = true
			if err := seedOne(next, product); err != nil {
				return err
			}
			if _, exists := next.Catalog[product.ID]; !exists {
				inserted++
				next.Catalog[product.ID] = product
			}
		}
		return nil
	})
	return inserted, err
}

func validateSeedProduct(product Product) error {
	if product.ID == "" || strings.TrimSpace(product.Name) == "" ||
		product.PriceCents < 0 || product.Revision < 1 {
		return fmt.Errorf("%w: seed products need an ID, name, non-negative price, and positive revision", errInvalid)
	}
	return nil
}

func seedOne(next *state, product Product) error {
	if existing, exists := next.Catalog[product.ID]; exists && existing != product {
		return fmt.Errorf("%w: seed product %q conflicts with existing catalog data", errConflict, product.ID)
	}
	return nil
}
