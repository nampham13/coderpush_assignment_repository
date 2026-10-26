package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (s *Store) ApplyBatch(updateID string, batch incomingBatch) (*Update, bool, error) {
	if strings.TrimSpace(batch.ID) == "" {
		return nil, false, fmt.Errorf("%w: batch id is required", errInvalid)
	}
	canonical, err := json.Marshal(batch)
	if err != nil {
		return nil, false, fmt.Errorf("encode batch: %w", err)
	}
	hash := sha256.Sum256(canonical)
	digest := hex.EncodeToString(hash[:])

	var result *Update
	idempotent := false
	err = s.change(func(next *state) error {
		update, ok := next.Updates[updateID]
		if !ok {
			return errNotFound
		}
		if update.Status != "in_progress" {
			return fmt.Errorf("%w: update is %s; resume it before submitting batches", errConflict, update.Status)
		}
		if previous, exists := update.Batches[batch.ID]; exists {
			if previous.Digest == digest {
				idempotent = true
				result = update
				return nil
			}
			addIssue(update, "batch_id_conflict", "Batch ID was reused with different content; the new content was not applied.", batch.ID, "", time.Now().UTC())
			update.UpdatedAt = time.Now().UTC()
			result = update
			return errBatchConflict
		}

		now := time.Now().UTC()
		summary := applyProducts(next, update, batch, now)
		summary.ID, summary.Digest, summary.Received = batch.ID, digest, now
		update.Batches[batch.ID] = summary
		update.UpdatedAt = now
		result = update
		return nil
	})
	return cloneUpdate(result), idempotent, err
}

func applyProducts(next *state, update *Update, batch incomingBatch, now time.Time) BatchSummary {
	summary := BatchSummary{Changes: []Change{}}
	for _, incoming := range batch.Products {
		id := strings.TrimSpace(incoming.ID)
		if invalidProduct(incoming, id) {
			summary.Rejected++
			addIssue(update, "invalid_product", "Product requires an ID, a positive revision, and (unless withdrawn) a name and non-negative price_cents.", batch.ID, id, now)
			continue
		}
		incoming.ID = id
		current, exists := next.Catalog[id]
		if exists && incoming.Revision < current.Revision {
			summary.Stale++
			continue
		}
		product := productFrom(incoming, current, exists)
		if exists && incoming.Revision == current.Revision {
			recordEqualRevision(update, &summary, product, current, batch.ID, id, now)
			continue
		}
		applyProductChange(next, update, &summary, product, current, exists)
	}
	return summary
}

func invalidProduct(product incomingProduct, id string) bool {
	return id == "" || product.Revision < 1 ||
		(!product.Withdrawn && (strings.TrimSpace(product.Name) == "" || product.PriceCents < 0))
}
