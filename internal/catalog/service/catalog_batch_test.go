package service

import (
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
