package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"renwen/server/internal/discussion"
	"renwen/server/internal/idgen"
	"renwen/server/internal/prompt"
	"renwen/server/internal/runtime"
	"renwen/server/internal/storage"
)

type GroupDiscussHandler struct {
	app *runtime.App
}

type groupDiscussRequest struct {
	ConversationID    string                   `json:"conversationId"`
	CharacterIDs      []string                 `json:"characterIds"`
	Question          string                   `json:"question"`
	SourceTitle       string                   `json:"sourceTitle"`
	SourceExcerpt     string                   `json:"sourceExcerpt"`
	SourceDetail      string                   `json:"sourceDetail"`
	HotURL            string                   `json:"hotUrl"`
	Provider          string                   `json:"provider"`
	History           []discussion.HistoryItem `json:"history"`
	Round             int                      `json:"round"`
	SpeakerIndex      *int                     `json:"speakerIndex,omitempty"`
	PriorTurnsInRound []discussion.Turn        `json:"priorTurnsInRound,omitempty"`
}

type groupDiscussResponse struct {
	ConversationID string            `json:"conversationId"`
	Turns          []discussion.Turn `json:"turns"`
	Provider       string            `json:"provider"`
	Model          string            `json:"model"`
}

func NewGroupDiscussHandler(app *runtime.App) *GroupDiscussHandler {
	return &GroupDiscussHandler{app: app}
}

func (h *GroupDiscussHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req groupDiscussRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	convID := strings.TrimSpace(req.ConversationID)
	if convID == "" {
		convID = idgen.New()
	}
	round := req.Round
	if round <= 0 {
		round = 1
	}

	orch, err := h.app.Orchestrator()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	source := prompt.SourceContext{
		Title:   req.SourceTitle,
		Excerpt: req.SourceExcerpt,
		Detail:  req.SourceDetail,
		URL:     req.HotURL,
	}

	turns, provider, model, err := orch.RunSpeaker(
		r.Context(), req.CharacterIDs, req.Question, source,
		req.Provider, req.History, round, speakerIndex(req.SpeakerIndex), req.PriorTurnsInRound,
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	store := h.app.Store
	saveUserMsg := req.SpeakerIndex == nil || (req.SpeakerIndex != nil && *req.SpeakerIndex == 0)
	if err := store.EnsureConversation(convID, "group", req.SourceTitle, req.HotURL, provider, req.CharacterIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if saveUserMsg {
		if _, err := store.AddMessage(storage.Message{
			ConversationID: convID, Role: "user", Content: strings.TrimSpace(req.Question), Round: round,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	for _, t := range turns {
		if _, err := store.AddMessage(storage.Message{
			ConversationID: convID, Role: "assistant", CharacterID: t.CharacterID,
			CharacterName: t.Name, Era: t.Era, Content: t.Content,
			Provider: provider, Model: model, Round: t.Round,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	_ = store.TouchConversation(convID)

	writeJSON(w, http.StatusOK, groupDiscussResponse{
		ConversationID: convID,
		Turns:          turns,
		Provider:       provider,
		Model:          model,
	})
}

func speakerIndex(v *int) int {
	if v == nil {
		return -1
	}
	return *v
}
