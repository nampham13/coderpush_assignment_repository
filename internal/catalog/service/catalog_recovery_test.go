package service

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestRestartPreservesPartialWorkAndRequiresExplicitCompletion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close original store: %v", err)
		}
	})
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
	t.Cleanup(func() {
		if err := restarted.Close(); err != nil {
			t.Errorf("close restarted store: %v", err)
		}
	})
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
