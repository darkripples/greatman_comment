package prompt

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// cacheListMeta mirrors handler.cacheListMeta for isolated testing.
type cacheListMeta struct {
	Cached      bool
	Stale       bool
	FetchedAt   time.Time
	ExpiresAt   time.Time
	MinInterval time.Duration
}

func writeCacheListResponse(w http.ResponseWriter, payload string, meta cacheListMeta) {
	var items any
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		http.Error(w, "invalid cache payload", http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []any{}
	}
	nextFetch := meta.FetchedAt.Add(meta.MinInterval)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items":       items,
		"cached":      meta.Cached,
		"stale":       meta.Stale,
		"source":      "sqlite",
		"fetchedAt":   meta.FetchedAt.Unix(),
		"expiresAt":   meta.ExpiresAt.Unix(),
		"nextFetchAt": nextFetch.Unix(),
	})
}

func TestWriteCacheListResponse(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()
	writeCacheListResponse(w, `["a","b"]`, cacheListMeta{
		Cached:      true,
		Stale:       false,
		FetchedAt:   now,
		ExpiresAt:   now.Add(5 * time.Minute),
		MinInterval: time.Minute,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["cached"] != true {
		t.Fatalf("expected cached=true, got %v", body["cached"])
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 items, got %v", body["items"])
	}
}
