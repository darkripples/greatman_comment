package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"renwen/server/internal/character"
	"renwen/server/internal/config"
	"renwen/server/internal/discussion"
	"renwen/server/internal/runtime"
	"renwen/server/internal/storage"
)

type ConversationsHandler struct {
	app *runtime.App
}

func NewConversationsHandler(app *runtime.App) *ConversationsHandler {
	return &ConversationsHandler{app: app}
}

func (h *ConversationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/conversations/")
	path = strings.Trim(path, "/")

	switch {
	case r.URL.Path == "/api/conversations" && r.Method == http.MethodGet:
		h.list(w, r)
	case path != "" && !strings.Contains(path, "/") && r.Method == http.MethodGet:
		h.get(w, path)
	case strings.HasSuffix(r.URL.Path, "/summarize") && r.Method == http.MethodPost:
		id := strings.TrimSuffix(path, "/summarize")
		h.summarize(w, r, id)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *ConversationsHandler) list(w http.ResponseWriter, r *http.Request) {
	limit := config.ParseLimit(r.URL.Query().Get("limit"), 20, 50)
	items, err := h.app.Store.ListConversations(limit)
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
	conv, msgs, err := h.app.Store.GetConversation(id)
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
	summary, _ := h.app.Store.GetConversationSummary(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"conversation": conv,
		"messages":     msgs,
		"summary":      summary,
	})
}

func (h *ConversationsHandler) summarize(w http.ResponseWriter, r *http.Request, id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}

	var body struct {
		Provider string `json:"provider"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	conv, msgs, err := h.app.Store.GetConversation(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if conv.Mode != "group" {
		writeError(w, http.StatusBadRequest, "only group conversations can be summarized")
		return
	}

	cfg := h.app.Config()
	provider := strings.TrimSpace(body.Provider)
	if provider == "" {
		provider = cfg.LLMProvider
	}

	content, err := discussion.GenerateRoundtableSummary(r.Context(), h.app.Router(), provider, conv, msgs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.app.Store.SaveConversationSummary(id, content); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conversationId": id,
		"content":        content,
	})
}
