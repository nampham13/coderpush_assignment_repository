package service

import (
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStoreFactory(filepath.Join(t.TempDir(), "catalog.db")).Create()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	return store
}
