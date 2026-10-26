package catalog

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStoreFactory(filepath.Join(t.TempDir(), "catalog.json")).Create()
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestFactoriesCreateApplicationComponents(t *testing.T) {
	store := openTestStore(t)
	handler := NewHandlerFactory(store).Create()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("factory-created handler returned %d", response.Code)
	}
}
