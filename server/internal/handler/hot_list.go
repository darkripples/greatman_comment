package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"renwen/server/internal/config"
	"renwen/server/internal/runtime"
)

type HotListHandler struct {
	app *runtime.App
}

func NewHotListHandler(app *runtime.App) *HotListHandler {
	return &HotListHandler{app: app}
}

func (h *HotListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	limit := config.ParseLimit(r.URL.Query().Get("limit"), 20, 30)
	key := fmt.Sprintf("hot_list:limit=%d", limit)
	store := h.app.Store

	if _, ok := store.GetAPICacheAny(key); !ok {
		_ = store.EnsureHotListSeedForKey(key)
	}

	entry, ok := store.GetAPICacheAny(key)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}

	var items any
	if err := json.Unmarshal([]byte(entry.Payload), &items); err != nil {
		writeError(w, http.StatusInternalServerError, "invalid hot list payload")
		return
	}
	if items == nil {
		items = []any{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
