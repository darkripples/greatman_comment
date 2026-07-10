package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"renwen/server/internal/storage"
)

type cacheListMeta struct {
	Cached      bool
	Stale       bool
	FetchedAt   time.Time
	ExpiresAt   time.Time
	MinInterval time.Duration
}

func serveCachedJSONList(
	w http.ResponseWriter,
	ctx context.Context,
	store *storage.Store,
	key string,
	force bool,
	ttl time.Duration,
	minInterval time.Duration,
	fetch func(context.Context) (any, error),
) {
	if entry, ok := store.GetAPICache(key); ok && !force {
		writeCacheListResponse(w, entry.Payload, cacheListMeta{
			Cached: true, Stale: false,
			FetchedAt: entry.FetchedAt, ExpiresAt: entry.ExpiresAt, MinInterval: minInterval,
		})
		return
	}

	if entry, ok := store.GetAPICacheAny(key); ok {
		if !force && time.Since(entry.FetchedAt) < minInterval {
			writeCacheListResponse(w, entry.Payload, cacheListMeta{
				Cached: true, Stale: true,
				FetchedAt: entry.FetchedAt, ExpiresAt: entry.ExpiresAt, MinInterval: minInterval,
			})
			return
		}
	}

	data, err := fetch(ctx)
	if err != nil {
		if entry, ok := store.GetAPICacheAny(key); ok {
			writeCacheListResponse(w, entry.Payload, cacheListMeta{
				Cached: true, Stale: true,
				FetchedAt: entry.FetchedAt, ExpiresAt: entry.ExpiresAt, MinInterval: minInterval,
			})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	raw, err := json.Marshal(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := store.SetAPICache(key, string(raw), ttl); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	entry, ok := store.GetAPICacheAny(key)
	if !ok {
		writeError(w, http.StatusInternalServerError, "cache write failed")
		return
	}
	writeCacheListResponse(w, entry.Payload, cacheListMeta{
		Cached: false, Stale: false,
		FetchedAt: entry.FetchedAt, ExpiresAt: entry.ExpiresAt, MinInterval: minInterval,
	})
}

func writeCacheListResponse(w http.ResponseWriter, payload string, meta cacheListMeta) {
	var items any
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		writeError(w, http.StatusInternalServerError, "invalid cache payload")
		return
	}
	if items == nil {
		items = []any{}
	}
	nextFetch := meta.FetchedAt.Add(meta.MinInterval)
	writeJSON(w, http.StatusOK, map[string]any{
		"items":       items,
		"cached":      meta.Cached,
		"stale":       meta.Stale,
		"source":      "sqlite",
		"fetchedAt":   meta.FetchedAt.Unix(),
		"expiresAt":   meta.ExpiresAt.Unix(),
		"nextFetchAt": nextFetch.Unix(),
	})
}
