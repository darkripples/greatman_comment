package handler

import (
	"net/http"

	"renwen/server/internal/character"
)

type CharactersHandler struct {
	store *character.Store
}

func NewCharactersHandler(store *character.Store) *CharactersHandler {
	return &CharactersHandler{store: store}
}

func (h *CharactersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": h.store.List()})
}
