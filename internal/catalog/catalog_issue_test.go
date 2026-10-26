package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRejectedProductBlocksCompletionUntilIssueIsAddressed(t *testing.T) {
	store := openTestStore(t)
	update, err := store.CreateUpdate("partial-batch")
	if err != nil {
		t.Fatal(err)
	}
	update, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "mixed", Products: []incomingProduct{
		{ID: "GOOD-1", Name: "Valid product", PriceCents: 100, Revision: 1},
		{ID: "BAD-1", Name: "Invalid product", PriceCents: -1, Revision: 1},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Catalog(false)) != 1 || len(update.Issues) != 1 {
		t.Fatalf("valid records should apply while invalid data remains visible: catalog=%#v issues=%#v", store.Catalog(false), update.Issues)
	}
	if _, err := store.CompleteUpdate(update.ID); !errors.Is(err, errConflict) {
		t.Fatalf("unresolved rejected product must block completion: %v", err)
	}
	update, err = store.ResolveIssue(update.ID, update.Issues[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	update, err = store.CompleteUpdate(update.ID)
	if err != nil || update.Status != "completed" {
		t.Fatalf("completion after issue resolution: update=%#v err=%v", update, err)
	}
}

func TestBatchIDConflictDoesNotApplyDifferentContent(t *testing.T) {
	store := openTestStore(t)
	update, err := store.CreateUpdate("conflicting-batch")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "same-id", Products: []incomingProduct{
		{ID: "ONE", Name: "Original", PriceCents: 100, Revision: 1},
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "same-id", Products: []incomingProduct{
		{ID: "ONE", Name: "Altered", PriceCents: 900, Revision: 2},
	}})
	if !errors.Is(err, errBatchConflict) {
		t.Fatalf("expected batch ID conflict, got %v", err)
	}
	product := store.Catalog(false)[0]
	if product.Name != "Original" || product.Revision != 1 {
		t.Fatalf("conflicting payload changed the catalog: %#v", product)
	}
}

func TestDemoEndpointShowsChangesAndWithdrawals(t *testing.T) {
	store := openTestStore(t)
	server := httptest.NewServer(NewHandler(store))
	defer server.Close()
	response, err := http.Post(server.URL+"/demo", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("demo status: %s", response.Status)
	}
	var update Update
	if err := json.NewDecoder(response.Body).Decode(&update); err != nil {
		t.Fatal(err)
	}
	if update.Status != "completed" || len(update.Changes) != 3 || len(update.Batches) != 2 {
		t.Fatalf("unexpected demo update: %#v", update)
	}
	if update.Batches["supplier-batch-01"].Stale != 1 {
		t.Fatalf("demo should include out-of-order stale data: %#v", update.Batches)
	}
}
