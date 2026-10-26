package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PriceCents  int64  `json:"price_cents"`
	Revision    int64  `json:"revision"`
	Withdrawn   bool   `json:"withdrawn"`
}

type Change struct {
	ProductID string   `json:"product_id"`
	Kind      string   `json:"kind"`
	Before    *Product `json:"before,omitempty"`
	After     Product  `json:"after"`
}

type Issue struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	BatchID   string    `json:"batch_id,omitempty"`
	ProductID string    `json:"product_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Resolved  bool      `json:"resolved"`
}

type BatchSummary struct {
	ID        string    `json:"id"`
	Digest    string    `json:"digest"`
	Received  time.Time `json:"received_at"`
	Applied   int       `json:"applied"`
	Stale     int       `json:"stale"`
	Unchanged int       `json:"unchanged"`
	Rejected  int       `json:"rejected"`
	Changes   []Change  `json:"changes"`
}

type Update struct {
	ID          string                  `json:"id"`
	Status      string                  `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	Batches     map[string]BatchSummary `json:"batches"`
	Changes     []Change                `json:"changes"`
	Issues      []Issue                 `json:"issues"`
}

type incomingProduct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Revision    int64  `json:"revision"`
	Withdrawn   bool   `json:"withdrawn"`
}

type incomingBatch struct {
	ID       string            `json:"id"`
	Products []incomingProduct `json:"products"`
}

type state struct {
	Catalog map[string]Product `json:"catalog"`
	Updates map[string]*Update `json:"updates"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	data state
}

func OpenStore(path string) (*Store, error) {
	s := &Store{path: path, data: state{Catalog: map[string]Product{}, Updates: map[string]*Update{}}}
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read catalog database: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(raw, &s.data); err != nil {
			return nil, fmt.Errorf("decode catalog database: %w", err)
		}
		if s.data.Catalog == nil {
			s.data.Catalog = map[string]Product{}
		}
		if s.data.Updates == nil {
			s.data.Updates = map[string]*Update{}
		}
	}
	for _, update := range s.data.Updates {
		if update.Status == "in_progress" {
			update.Status = "interrupted"
			update.UpdatedAt = time.Now().UTC()
		}
	}
	if err := s.persist(s.data); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) CreateUpdate(id string) (*Update, error) {
	if strings.TrimSpace(id) == "" {
		generated, err := newID()
		if err != nil {
			return nil, err
		}
		id = generated
	}
	now := time.Now().UTC()
	var result *Update
	err := s.change(func(next *state) error {
		if existing, ok := next.Updates[id]; ok {
			result = existing
			return nil
		}
		result = &Update{
			ID: id, Status: "in_progress", CreatedAt: now, UpdatedAt: now,
			Batches: map[string]BatchSummary{}, Changes: []Change{}, Issues: []Issue{},
		}
		next.Updates[id] = result
		return nil
	})
	return cloneUpdate(result), err
}

func (s *Store) ResumeUpdate(id string) (*Update, error) {
	var result *Update
	err := s.change(func(next *state) error {
		update, ok := next.Updates[id]
		if !ok {
			return errNotFound
		}
		if update.Status == "completed" {
			return fmt.Errorf("%w: completed updates cannot be resumed", errConflict)
		}
		update.Status = "in_progress"
		update.UpdatedAt = time.Now().UTC()
		result = update
		return nil
	})
	return cloneUpdate(result), err
}

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
		summary := BatchSummary{ID: batch.ID, Digest: digest, Received: now, Changes: []Change{}}
		for _, incoming := range batch.Products {
			id := strings.TrimSpace(incoming.ID)
			if id == "" || incoming.Revision < 1 || (!incoming.Withdrawn && (strings.TrimSpace(incoming.Name) == "" || incoming.PriceCents < 0)) {
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
			product := Product{
				ID: id, Name: incoming.Name, Description: incoming.Description,
				PriceCents: incoming.PriceCents, Revision: incoming.Revision, Withdrawn: incoming.Withdrawn,
			}
			if exists && incoming.Withdrawn {
				product.Name = current.Name
				product.Description = current.Description
				product.PriceCents = current.PriceCents
			}
			if exists && incoming.Revision == current.Revision {
				if product == current {
					summary.Unchanged++
				} else {
					summary.Rejected++
					addIssue(update, "revision_conflict", "Product data differs from the catalog at the same supplier revision; the existing product was kept.", batch.ID, id, now)
				}
				continue
			}
			kind := "created"
			var before *Product
			switch {
			case !exists:
				kind = "created"
			case current.Withdrawn && !product.Withdrawn:
				kind = "restored"
				copy := current
				before = &copy
			case !current.Withdrawn && product.Withdrawn:
				kind = "withdrawn"
				copy := current
				before = &copy
			default:
				kind = "updated"
				copy := current
				before = &copy
			}
			next.Catalog[id] = product
			change := Change{ProductID: id, Kind: kind, Before: before, After: product}
			summary.Applied++
			summary.Changes = append(summary.Changes, change)
			update.Changes = append(update.Changes, change)
		}
		update.Batches[batch.ID] = summary
		update.UpdatedAt = now
		result = update
		return nil
	})
	return cloneUpdate(result), idempotent, err
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

var (
	errNotFound      = errors.New("not found")
	errBatchConflict = errors.New("batch ID conflict")
	errConflict      = errors.New("conflict")
	errInvalid       = errors.New("invalid request")
)

func (s *Store) change(mutate func(*state) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := json.Marshal(s.data)
	if err != nil {
		return fmt.Errorf("copy catalog state: %w", err)
	}
	var next state
	if err := json.Unmarshal(raw, &next); err != nil {
		return fmt.Errorf("copy catalog state: %w", err)
	}
	if err := mutate(&next); err != nil {
		if !errors.Is(err, errBatchConflict) {
			return err
		}
		if saveErr := s.persist(next); saveErr != nil {
			return saveErr
		}
		s.data = next
		return err
	}
	if err := s.persist(next); err != nil {
		return err
	}
	s.data = next
	return nil
}

func (s *Store) persist(next state) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create catalog database directory: %w", err)
	}
	raw, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return fmt.Errorf("encode catalog database: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(s.path), ".catalog-*.tmp")
	if err != nil {
		return fmt.Errorf("create catalog database temp file: %w", err)
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("secure catalog database: %w", err)
	}
	if _, err := file.Write(raw); err != nil {
		file.Close()
		return fmt.Errorf("write catalog database: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync catalog database: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close catalog database: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace catalog database: %w", err)
	}
	return nil
}

func addIssue(update *Update, code, message, batchID, productID string, now time.Time) {
	for _, issue := range update.Issues {
		if issue.Code == code && issue.Message == message && issue.BatchID == batchID && issue.ProductID == productID && !issue.Resolved {
			return
		}
	}
	id := fmt.Sprintf("issue-%d-%d", now.UnixNano(), len(update.Issues)+1)
	update.Issues = append(update.Issues, Issue{
		ID: id, Code: code, Message: message, BatchID: batchID, ProductID: productID, CreatedAt: now,
	})
}

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate ID: %w", err)
	}
	return hex.EncodeToString(bytes[:]), nil
}

func cloneUpdate(update *Update) *Update {
	if update == nil {
		return nil
	}
	raw, _ := json.Marshal(update)
	var copy Update
	_ = json.Unmarshal(raw, &copy)
	return &copy
}
