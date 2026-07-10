package handler

import (
	"net/http"
	"strings"

	"renwen/server/internal/character"
	"renwen/server/internal/config"
	"renwen/server/internal/storage"
)

type ConversationsHandler struct {
	store *storage.Store
}

func NewConversationsHandler(store *storage.Store) *ConversationsHandler {
	return &ConversationsHandler{store: store}
}

func (h *ConversationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api/conversations" && r.Method == http.MethodGet:
		h.list(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/conversations/") && r.Method == http.MethodGet:
		id := strings.TrimPrefix(r.URL.Path, "/api/conversations/")
		h.get(w, id)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *ConversationsHandler) list(w http.ResponseWriter, r *http.Request) {
	limit := config.ParseLimit(r.URL.Query().Get("limit"), 20, 50)
	items, err := h.store.ListConversations(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []storage.Conversation{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ConversationsHandler) get(w http.ResponseWriter, id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	conv, msgs, err := h.store.GetConversation(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if msgs == nil {
		msgs = []storage.Message{}
	}
	for i := range msgs {
		if msgs[i].Role == "assistant" {
			msgs[i].Content = character.SanitizeReply(msgs[i].Content)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"conversation": conv,
		"messages":     msgs,
	})
}
