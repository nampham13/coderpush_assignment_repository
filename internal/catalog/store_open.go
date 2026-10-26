package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

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
