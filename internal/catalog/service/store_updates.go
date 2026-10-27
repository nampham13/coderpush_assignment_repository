package service

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func (s *Store) RecordOutage(id, message string) (*Update, error) {
	var result *Update
	err := s.change(func(next *state) error {
		update, ok := next.Updates[id]
		if !ok {
			return errNotFound
		}
		if update.Status != "in_progress" {
			return fmt.Errorf("%w: update is %s; resume it before recording an outage", errConflict, update.Status)
		}
		if strings.TrimSpace(message) == "" {
			return fmt.Errorf("%w: message is required", errInvalid)
		}
		now := time.Now().UTC()
		addIssue(update, "supplier_outage", strings.TrimSpace(message), "", "", now)
		update.Status = "interrupted"
		update.UpdatedAt = now
		result = update
		return nil
	})
	return cloneUpdate(result), err
}

func (s *Store) ResolveIssue(updateID, issueID string) (*Update, error) {
	var result *Update
	err := s.change(func(next *state) error {
		update, ok := next.Updates[updateID]
		if !ok {
			return errNotFound
		}
		for i := range update.Issues {
			if update.Issues[i].ID == issueID {
				update.Issues[i].Resolved = true
				update.UpdatedAt = time.Now().UTC()
				result = update
				return nil
			}
		}
		return errNotFound
	})
	return cloneUpdate(result), err
}

func (s *Store) Catalog(includeWithdrawn bool) []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	products := make([]Product, 0, len(s.data.Catalog))
	for _, product := range s.data.Catalog {
		if includeWithdrawn || !product.Withdrawn {
			products = append(products, product)
		}
	}
	sort.Slice(products, func(i, j int) bool { return products[i].ID < products[j].ID })
	return products
}

func (s *Store) GetUpdate(id string) (*Update, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	update, ok := s.data.Updates[id]
	if !ok {
		return nil, errNotFound
	}
	return cloneUpdate(update), nil
}

func (s *Store) Updates() []Update {
	s.mu.RLock()
	defer s.mu.RUnlock()
	updates := make([]Update, 0, len(s.data.Updates))
	for _, update := range s.data.Updates {
		updates = append(updates, *cloneUpdate(update))
	}
	sort.Slice(updates, func(i, j int) bool { return updates[i].CreatedAt.After(updates[j].CreatedAt) })
	return updates
}
