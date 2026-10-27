package httpapi

import (
	"net/http"

	"example.com/supplier-catalog/internal/catalog/domain"
	"example.com/supplier-catalog/internal/catalog/service"
)

type API struct {
	store *service.Store
}

func NewHandler(store *service.Store) http.Handler {
	return NewHandlerFactory(store).Create()
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
	var batch domain.IncomingBatch
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
