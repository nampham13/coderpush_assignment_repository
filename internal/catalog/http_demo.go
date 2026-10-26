package catalog

import (
	"net/http"
)

func (api *API) demo(w http.ResponseWriter, _ *http.Request) {
	baseline, err := api.store.CreateUpdate("demo-baseline")
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if baseline.Status != "completed" {
		if baseline.Status == "interrupted" {
			baseline, err = api.store.ResumeUpdate(baseline.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
		}
		_, _, err = api.store.ApplyBatch(baseline.ID, incomingBatch{
			ID: "baseline-1",
			Products: []incomingProduct{
				{ID: "TEA-100", Name: "Harbor Breakfast Tea", Description: "Loose-leaf black tea", PriceCents: 1299, Revision: 1},
				{ID: "MUG-200", Name: "Kiln Stoneware Mug", Description: "Hand-finished ceramic mug", PriceCents: 1899, Revision: 1},
			},
		})
		if err == nil {
			_, err = api.store.CompleteUpdate(baseline.ID)
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
	}

	update, err := api.store.CreateUpdate("demo-spring-2026")
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if update.Status == "completed" {
		writeJSON(w, http.StatusOK, update)
		return
	}
	if update.Status == "interrupted" {
		update, err = api.store.ResumeUpdate(update.ID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
	}
	_, _, err = api.store.ApplyBatch(update.ID, incomingBatch{
		ID: "supplier-batch-02",
		Products: []incomingProduct{
			{ID: "TEA-100", Name: "Harbor Breakfast Tea", Description: "Updated seasonal blend", PriceCents: 1399, Revision: 3},
			{ID: "MUG-200", Revision: 2, Withdrawn: true},
		},
	})
	if err == nil {
		_, _, err = api.store.ApplyBatch(update.ID, incomingBatch{
			ID: "supplier-batch-01",
			Products: []incomingProduct{
				{ID: "TEA-100", Name: "Older tea details", PriceCents: 1199, Revision: 2},
				{ID: "COF-300", Name: "North Pier Coffee", Description: "Whole-bean medium roast", PriceCents: 1599, Revision: 1},
			},
		})
	}
	if err == nil {
		update, err = api.store.CompleteUpdate(update.ID)
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, update)
}
