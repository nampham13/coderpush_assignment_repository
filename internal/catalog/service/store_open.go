package service

import (
	"fmt"
	"strings"
	"time"

	"example.com/supplier-catalog/internal/catalog/repository/sqlite"
)

func OpenStore(path string) (*Store, error) {
	repository, err := sqlite.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := repository.Load()
	if err != nil {
		repository.Close()
		return nil, err
	}
	s := &Store{repo: repository, data: data}
	for _, update := range s.data.Updates {
		if update.Status == "in_progress" {
			update.Status = "interrupted"
			update.UpdatedAt = time.Now().UTC()
		}
	}
	if err := s.repo.Save(s.data); err != nil {
		s.Close()
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
			return fmt.Errorf("%w: update ID %q already exists with status %s", errConflict, id, existing.Status)
		}
		if active := activeUpdate(next, ""); active != nil {
			return fmt.Errorf("%w: update %q is still %s", errConflict, active.ID, active.Status)
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
		if active := activeUpdate(next, id); active != nil {
			return fmt.Errorf("%w: update %q is still %s", errConflict, active.ID, active.Status)
		}
		update.Status = "in_progress"
		update.UpdatedAt = time.Now().UTC()
		result = update
		return nil
	})
	return cloneUpdate(result), err
}

func activeUpdate(data *state, exceptID string) *Update {
	for id, update := range data.Updates {
		if id != exceptID && (update.Status == "in_progress" || update.Status == "interrupted") {
			return update
		}
	}
	return nil
}

func (s *Store) Close() error {
	return s.repo.Close()
}
