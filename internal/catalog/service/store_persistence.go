package service

import (
	"encoding/json"
	"errors"
	"fmt"
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
	mutationErr := mutate(&next)
	if mutationErr != nil && !errors.Is(mutationErr, errBatchConflict) {
		return mutationErr
	}
	if err := s.repo.Save(next); err != nil {
		return err
	}
	s.data = next
	return mutationErr
}
