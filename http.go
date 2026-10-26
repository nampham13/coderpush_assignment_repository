package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type API struct {
	store *Store
}

func NewHandler(store *Store) http.Handler {
	api := &API{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.health)
	mux.HandleFunc("GET /catalog", api.catalog)
	mux.HandleFunc("GET /updates", api.updates)
	mux.HandleFunc("POST /updates", api.createUpdate)
	mux.HandleFunc("POST /demo", api.demo)
	mux.HandleFunc("GET /updates/{updateID}", api.getUpdate)
	mux.HandleFunc("POST /updates/{updateID}/resume", api.resumeUpdate)
	mux.HandleFunc("POST /updates/{updateID}/batches", api.applyBatch)
	mux.HandleFunc("POST /updates/{updateID}/complete", api.completeUpdate)
	mux.HandleFunc("POST /updates/{updateID}/failures", api.recordOutage)
	mux.HandleFunc("POST /updates/{updateID}/issues/{issueID}/resolve", api.resolveIssue)
	return mux
}

func (api *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *API) catalog(w http.ResponseWriter, r *http.Request) {
	includeWithdrawn := r.URL.Query().Get("include_withdrawn") == "true"
	writeJSON(w, http.StatusOK, map[string]any{"products": api.store.Catalog(includeWithdrawn)})
}

func (api *API) updates(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"updates": api.store.Updates()})
}

func (api *API) createUpdate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID string `json:"id"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	update, err := api.store.CreateUpdate(request.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, update)
}

func (api *API) getUpdate(w http.ResponseWriter, r *http.Request) {
	update, err := api.store.GetUpdate(r.PathValue("updateID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, update)
}

func (api *API) resumeUpdate(w http.ResponseWriter, r *http.Request) {
	update, err := api.store.ResumeUpdate(r.PathValue("updateID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, update)
}

func (api *API) applyBatch(w http.ResponseWriter, r *http.Request) {
	var batch incomingBatch
	if err := decodeJSON(w, r, &batch); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	update, replayed, err := api.store.ApplyBatch(r.PathValue("updateID"), batch)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	status := http.StatusOK
	if replayed {
		w.Header().Set("Idempotent-Replay", "true")
	}
	writeJSON(w, status, update)
}

func (api *API) completeUpdate(w http.ResponseWriter, r *http.Request) {
	update, err := api.store.CompleteUpdate(r.PathValue("updateID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, update)
}

func (api *API) recordOutage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Message string `json:"message"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	update, err := api.store.RecordOutage(r.PathValue("updateID"), request.Message)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, update)
}

func (api *API) resolveIssue(w http.ResponseWriter, r *http.Request) {
	update, err := api.store.ResolveIssue(r.PathValue("updateID"), r.PathValue("issueID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, update)
}

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

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON request: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request must contain one JSON value")
		}
		return fmt.Errorf("invalid trailing JSON: %w", err)
	}
	return nil
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errNotFound):
		writeError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, errBatchConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, errInvalid):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, errConflict):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		fmt.Printf("write HTTP response: %v\n", err)
	}
}
