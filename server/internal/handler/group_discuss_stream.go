package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"renwen/server/internal/character"
	"renwen/server/internal/discussion"
	"renwen/server/internal/idgen"
	"renwen/server/internal/prompt"
	"renwen/server/internal/runtime"
	"renwen/server/internal/storage"
)

type GroupDiscussStreamHandler struct {
	app *runtime.App
}

func NewGroupDiscussStreamHandler(app *runtime.App) *GroupDiscussStreamHandler {
	return &GroupDiscussStreamHandler{app: app}
}

func (h *GroupDiscussStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req groupDiscussRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	sse, ok := newSSEWriter(w)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
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
		_ = sse.Event("error", map[string]string{"message": err.Error()})
		return
	}

	source := prompt.SourceContext{
		Title: req.SourceTitle, Excerpt: req.SourceExcerpt, Detail: req.SourceDetail, URL: req.HotURL,
	}

	_ = sse.Event("meta", map[string]string{"conversationId": convID})

	var completed []discussion.Turn
	var provider, model string

	// Stream one speaker at a time for progressive UI updates.
	characterIDs := req.CharacterIDs
	for i := range characterIDs {
		idx := i
		ch, ok := h.app.Characters.Get(characterIDs[i])
		if !ok {
			continue
		}
		_ = sse.Event("turn_start", map[string]any{
			"characterId": ch.ID,
			"name":        ch.Name,
			"era":         ch.Era,
			"round":       round,
			"index":       idx,
		})

		turns, prov, mdl, err := orch.RunSpeaker(
			r.Context(), characterIDs, req.Question, source,
			req.Provider, req.History, round, idx, completed,
		)
		provider = prov
		model = mdl
		if err != nil {
			_ = sse.Event("error", map[string]string{
				"message": err.Error(),
			})
			return
		}
		if len(turns) == 0 {
			_ = sse.Event("error", map[string]string{"message": "empty turn"})
			return
		}
		t := turns[0]
		completed = append(completed, t)

		store := h.app.Store
		if i == 0 {
			if err := store.EnsureConversation(convID, "group", req.SourceTitle, req.HotURL, provider, characterIDs); err != nil {
				_ = sse.Event("error", map[string]string{"message": err.Error()})
				return
			}
			if _, err := store.AddMessage(storage.Message{
				ConversationID: convID, Role: "user", Content: strings.TrimSpace(req.Question), Round: round,
			}); err != nil {
				_ = sse.Event("error", map[string]string{"message": err.Error()})
				return
			}
		}
		content := character.SanitizeReply(t.Content)
		if _, err := store.AddMessage(storage.Message{
			ConversationID: convID, Role: "assistant", CharacterID: t.CharacterID,
			CharacterName: t.Name, Era: t.Era, Content: content,
			Provider: provider, Model: model, Round: t.Round, Citations: citationsToStorage(t.Citations),
		}); err != nil {
			_ = sse.Event("error", map[string]string{"message": err.Error()})
			return
		}
		_ = store.TouchConversation(convID)

		_ = sse.Event("turn_done", map[string]any{
			"characterId": t.CharacterID,
			"name":        t.Name,
			"era":         t.Era,
			"content":     content,
			"round":       t.Round,
			"citations":   citationsToJSON(t.Citations),
		})
	}

	_ = sse.Event("done", map[string]any{
		"conversationId": convID,
		"provider":       provider,
		"model":          model,
		"turns":          completed,
	})
}
