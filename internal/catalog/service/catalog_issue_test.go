package service

import (
	"errors"
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
