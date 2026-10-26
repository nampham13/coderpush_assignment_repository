package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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
