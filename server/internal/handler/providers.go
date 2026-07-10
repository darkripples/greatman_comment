package handler

import (
	"net/http"

	"renwen/server/internal/runtime"
)

type ProvidersHandler struct {
	app *runtime.App
}

func NewProvidersHandler(app *runtime.App) *ProvidersHandler {
	return &ProvidersHandler{app: app}
}

func (h *ProvidersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": h.app.Router().List()})
}
