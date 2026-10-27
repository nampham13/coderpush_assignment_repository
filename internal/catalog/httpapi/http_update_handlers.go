package httpapi

import (
	"net/http"

	"example.com/supplier-catalog/internal/catalog/domain"
)

func (api *API) demoUpdate(id string) (*domain.Update, error) {
	update, err := api.store.GetUpdate(id)
	if err == nil {
		return update, nil
	}
	if err != domain.ErrNotFound {
		return nil, err
	}
	return api.store.CreateUpdate(id)
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
