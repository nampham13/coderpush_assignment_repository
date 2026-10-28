package seed

import (
	"path/filepath"
	"testing"

	"example.com/supplier-catalog/internal/catalog/service"
)

func TestImportMockDatabaseIsIdempotent(t *testing.T) {
	store, err := service.NewStoreFactory(filepath.Join(t.TempDir(), "catalog.db")).Create()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	source := filepath.Join("..", "..", "..", "examples", "mock-data.db")
	assertImportedCount(t, store, source, 100)
	assertImportedCount(t, store, source, 0)
	if count := len(store.Catalog(false)); count != 100 {
		t.Fatalf("seeded %d products, want 100", count)
	}
}

func assertImportedCount(t *testing.T, store *service.Store, source string, want int) {
	t.Helper()
	got, err := Import(store, source)
	if err != nil || got != want {
		t.Fatalf("Import() = %d, %v; want %d", got, err, want)
	}
}
