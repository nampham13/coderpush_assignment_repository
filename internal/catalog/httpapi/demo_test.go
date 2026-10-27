package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/supplier-catalog/internal/catalog/domain"
)

func TestDemoEndpointShowsChangesAndWithdrawals(t *testing.T) {
	server := httptest.NewServer(NewHandler(openHTTPTestStore(t)))
	defer server.Close()
	response, err := http.Post(server.URL+"/demo", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("demo status: %s", response.Status)
	}
	assertDemoUpdate(t, response)
}

func assertDemoUpdate(t *testing.T, response *http.Response) {
	var update domain.Update
	if err := json.NewDecoder(response.Body).Decode(&update); err != nil {
		t.Fatal(err)
	}
	if update.Status != "completed" || len(update.Changes) != 3 || len(update.Batches) != 2 {
		t.Fatalf("unexpected demo update: %#v", update)
	}
	if update.Batches["supplier-batch-01"].Stale != 1 {
		t.Fatalf("demo should include stale data: %#v", update.Batches)
	}
}
