package service

import (
	"fmt"
	"time"
)

func productFrom(in incomingProduct, current Product, exists bool) Product {
	product := Product{ID: in.ID, Name: in.Name, Description: in.Description,
		PriceCents: in.PriceCents, Revision: in.Revision, Withdrawn: in.Withdrawn}
	if exists && in.Withdrawn {
		product.Name, product.Description, product.PriceCents = current.Name, current.Description, current.PriceCents
	}
	return product
}

func recordEqualRevision(update *Update, summary *BatchSummary, product, current Product, batchID, id string, now time.Time) {
	if product == current {
		summary.Unchanged++
		return
	}
	summary.Rejected++
	addIssue(update, "revision_conflict", "Product data differs from the catalog at the same supplier revision; the existing product was kept.", batchID, id, now)
}

func applyProductChange(next *state, update *Update, summary *BatchSummary, product, current Product, exists bool) {
	kind, before := productChangeKind(product, current, exists)
	next.Catalog[product.ID] = product
	change := Change{ProductID: product.ID, Kind: kind, Before: before, After: product}
	summary.Applied++
	summary.Changes = append(summary.Changes, change)
	update.Changes = append(update.Changes, change)
}

func productChangeKind(product, current Product, exists bool) (string, *Product) {
	if !exists {
		return "created", nil
	}
	before := current
	switch {
	case current.Withdrawn && !product.Withdrawn:
		return "restored", &before
	case !current.Withdrawn && product.Withdrawn:
		return "withdrawn", &before
	default:
		return "updated", &before
	}
}

func (s *Store) CompleteUpdate(id string) (*Update, error) {
	var result *Update
	err := s.change(func(next *state) error {
		update, ok := next.Updates[id]
		if !ok {
			return errNotFound
		}
		if update.Status == "completed" {
			result = update
			return nil
		}
		if update.Status != "in_progress" {
			return fmt.Errorf("%w: update is %s; resume it before completing", errConflict, update.Status)
		}
		for _, issue := range update.Issues {
			if !issue.Resolved {
				return fmt.Errorf("%w: resolve all outstanding issues before completing the update", errConflict)
			}
		}
		now := time.Now().UTC()
		update.Status = "completed"
		update.UpdatedAt = now
		update.CompletedAt = &now
		result = update
		return nil
	})
	return cloneUpdate(result), err
}
