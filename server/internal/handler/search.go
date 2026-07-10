package handler

import (
	"context"
	"fmt"
	"net/http"

	"renwen/server/internal/config"
	"renwen/server/internal/runtime"
	"renwen/server/internal/zhihu"
)

type SearchHandler struct {
	app *runtime.App
}

func NewSearchHandler(app *runtime.App) *SearchHandler {
	return &SearchHandler{app: app}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cfg := h.app.Config()
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q is required")
		return
	}
	count := config.ParseLimit(r.URL.Query().Get("count"), 10, 10)
	key := fmt.Sprintf("search:q=%s:count=%d", q, count)
	force := r.URL.Query().Get("force") == "1"
	client := h.app.ZhihuClient()

	serveCachedJSONList(w, r.Context(), h.app.Store, key, force, cfg.SearchCacheTTL, cfg.SearchMinInterval,
		func(ctx context.Context) (any, error) {
			items, err := client.Search(ctx, q, count)
			if err != nil {
				return nil, err
			}
			if items == nil {
				items = []zhihu.SearchItem{}
			}
			return items, nil
		},
	)
}
