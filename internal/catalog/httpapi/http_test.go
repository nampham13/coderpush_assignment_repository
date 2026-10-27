package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"example.com/supplier-catalog/internal/catalog/service"
)

func openHTTPTestStore(t *testing.T) *service.Store {
	t.Helper()
	store, err := service.NewStoreFactory(filepath.Join(t.TempDir(), "catalog.db")).Create()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close HTTP test database: %v", err)
		}
	})
	return store
}

func TestHandlerFactoryRegistersHealthRoute(t *testing.T) {
	handler := NewHandlerFactory(openHTTPTestStore(t)).Create()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("factory-created handler returned %d", response.Code)
	}
}
