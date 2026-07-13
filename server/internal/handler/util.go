package handler

import (
	"encoding/json"
	"net/http"

	"renwen/server/internal/rag"
	"renwen/server/internal/storage"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func citationsToJSON(citations []rag.Citation) []map[string]string {
	if len(citations) == 0 {
		return nil
	}
	out := make([]map[string]string, 0, len(citations))
	for _, c := range citations {
		out = append(out, map[string]string{"title": c.Title, "source": c.Source, "excerpt": c.Excerpt})
	}
	return out
}

func citationsToStorage(citations []rag.Citation) []storage.Citation {
	if len(citations) == 0 {
		return nil
	}
	out := make([]storage.Citation, 0, len(citations))
	for _, c := range citations {
		out = append(out, storage.Citation{Title: c.Title, Source: c.Source, Excerpt: c.Excerpt})
	}
	return out
}
