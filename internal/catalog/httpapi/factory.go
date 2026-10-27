package httpapi

import (
	"net/http"

	"example.com/supplier-catalog/internal/catalog/service"
)

type HandlerFactory struct {
	store *service.Store
}

func NewHandlerFactory(store *service.Store) HandlerFactory {
	return HandlerFactory{store: store}
}

func (factory HandlerFactory) Create() http.Handler {
	api := &API{store: factory.store}
	mux := http.NewServeMux()
	registerRoutes(mux, api)
	return mux
}

func registerRoutes(mux *http.ServeMux, api *API) {
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
}
