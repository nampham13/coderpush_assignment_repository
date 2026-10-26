package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestBatchReplayStaleRevisionAndWithdrawal(t *testing.T) {
	store := openTestStore(t)
	update, err := store.CreateUpdate("supplier-run")
	if err != nil {
		t.Fatal(err)
	}
	first := incomingBatch{ID: "batch-a", Products: []incomingProduct{
		{ID: "TEA-1", Name: "Original", PriceCents: 500, Revision: 1},
	}}
	_, replayed, err := store.ApplyBatch(update.ID, first)
	if err != nil || replayed {
		t.Fatalf("first batch: replayed=%v err=%v", replayed, err)
	}
	_, replayed, err = store.ApplyBatch(update.ID, first)
	if err != nil || !replayed {
		t.Fatalf("batch replay: replayed=%v err=%v", replayed, err)
	}

	_, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "batch-b", Products: []incomingProduct{
		{ID: "TEA-1", Name: "Current", PriceCents: 700, Revision: 3},
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "batch-c", Products: []incomingProduct{
		{ID: "TEA-1", Name: "Stale", PriceCents: 100, Revision: 2},
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "batch-d", Products: []incomingProduct{
		{ID: "TEA-1", Revision: 4, Withdrawn: true},
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "batch-e", Products: []incomingProduct{
		{ID: "TEA-1", Name: "Must not restore", PriceCents: 999, Revision: 3},
	}})
	if err != nil {
		t.Fatal(err)
	}

	all := store.Catalog(true)
	if len(all) != 1 || !all[0].Withdrawn || all[0].Revision != 4 || all[0].Name != "Current" || all[0].PriceCents != 700 {
		t.Fatalf("catalog after stale update: %#v", all)
	}
	if active := store.Catalog(false); len(active) != 0 {
		t.Fatalf("withdrawn product was listed as active: %#v", active)
	}
	loaded, err := store.GetUpdate(update.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Changes) != 3 {
		t.Fatalf("expected exactly three real changes after replay/stale batch; got %d", len(loaded.Changes))
	}
	if loaded.Batches["batch-c"].Stale != 1 {
		t.Fatalf("stale batch summary: %#v", loaded.Batches["batch-c"])
	}
}

func TestRestartPreservesPartialWorkAndRequiresExplicitCompletion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.json")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	update, err := store.CreateUpdate("recoverable")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ApplyBatch(update.ID, incomingBatch{ID: "first", Products: []incomingProduct{
		{ID: "BOOK-1", Name: "Field Guide", PriceCents: 2500, Revision: 1},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetUpdate(update.ID); err != nil {
		t.Fatal(err)
	}

	restarted, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := restarted.GetUpdate(update.ID)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Status != "interrupted" || len(recovered.Changes) != 1 {
		t.Fatalf("restart state did not preserve partial work: %#v", recovered)
	}
	if _, err := restarted.CompleteUpdate(update.ID); !errors.Is(err, errConflict) {
		t.Fatalf("interrupted update should not complete: %v", err)
	}
	recovered, err = restarted.ResumeUpdate(update.ID)
	if err != nil || recovered.Status != "in_progress" {
		t.Fatalf("resume: update=%#v err=%v", recovered, err)
	}
	_, replayed, err := restarted.ApplyBatch(update.ID, incomingBatch{ID: "first", Products: []incomingProduct{
		{ID: "BOOK-1", Name: "Field Guide", PriceCents: 2500, Revision: 1},
	}})
	if err != nil || !replayed {
		t.Fatalf("replay after restart: replayed=%v err=%v", replayed, err)
	}
	recovered, err = restarted.CompleteUpdate(update.ID)
	if err != nil || recovered.Status != "completed" || recovered.CompletedAt == nil {
		t.Fatalf("completion: update=%#v err=%v", recovered, err)
	}
}

func TestOutageMustBeResolvedBeforeCompletion(t *testing.T) {
	store := openTestStore(t)
	update, err := store.CreateUpdate("outage-run")
	if err != nil {
		t.Fatal(err)
	}
	update, err = store.RecordOutage(update.ID, "Fictional supplier is temporarily unavailable")
	if err != nil || update.Status != "interrupted" {
		t.Fatalf("record outage: update=%#v err=%v", update, err)
	}
	if _, err := store.CompleteUpdate(update.ID); !errors.Is(err, errConflict) {
		t.Fatalf("outage must block completion: %v", err)
	}
	update, err = store.ResumeUpdate(update.ID)
	if err != nil {
		t.Fatal(err)
	}
	update, err = store.ResolveIssue(update.ID, update.Issues[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	update, err = store.CompleteUpdate(update.ID)
	if err != nil || update.Status != "completed" || !update.Issues[0].Resolved {
		t.Fatalf("resolved outage should allow completion: update=%#v err=%v", update, err)
	}
	if _, err := store.CompleteUpdate(update.ID); err != nil {
		t.Fatalf("repeated completion should be idempotent: %v", err)
	}
}

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

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := OpenStore(filepath.Join(t.TempDir(), "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	return store
}
