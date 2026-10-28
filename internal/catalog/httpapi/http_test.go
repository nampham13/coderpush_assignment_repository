package httpapi

import (
	"encoding/json"
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

func TestCatalogRejectsInvalidWithdrawnQuery(t *testing.T) {
	handler := NewHandlerFactory(openHTTPTestStore(t)).Create()
	for _, path := range []string{"/catalog?include_withdrawn=maybe", "/catalog?include_withdrawn=true&include_withdrawn=false"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("catalog returned %d, want %d", response.Code, http.StatusBadRequest)
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("expected standard JSON error, got %q", response.Body.String())
			}
		})
	}
}

func TestCatalogAcceptsBooleanWithdrawnQuery(t *testing.T) {
	handler := NewHandlerFactory(openHTTPTestStore(t)).Create()
	request := httptest.NewRequest(http.MethodGet, "/catalog?include_withdrawn=true", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("catalog returned %d, want %d", response.Code, http.StatusOK)
	}
}
