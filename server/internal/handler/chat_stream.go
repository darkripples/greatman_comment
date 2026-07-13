package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"renwen/server/internal/character"
	"renwen/server/internal/discussion"
	"renwen/server/internal/idgen"
	"renwen/server/internal/llm"
	"renwen/server/internal/prompt"
	"renwen/server/internal/runtime"
	"renwen/server/internal/storage"
)

type ChatStreamHandler struct {
	app *runtime.App
}

func NewChatStreamHandler(app *runtime.App) *ChatStreamHandler {
	return &ChatStreamHandler{app: app}
}

func (h *ChatStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Question = strings.TrimSpace(req.Question)
	req.CharacterID = strings.TrimSpace(req.CharacterID)
	if req.CharacterID == "" || req.Question == "" {
		writeError(w, http.StatusBadRequest, "characterId and question are required")
		return
	}

	ch, ok := h.app.Characters.Get(req.CharacterID)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown characterId")
		return
	}

	round := req.Round
	if round <= 0 {
		round = 1
	}
	if round > discussion.MaxRound {
		writeError(w, http.StatusBadRequest, "round exceeds max")
		return
	}

	sse, ok := newSSEWriter(w)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	cfg := h.app.Config()
	convID := strings.TrimSpace(req.ConversationID)
	if convID == "" {
		convID = idgen.New()
	}
	providerName := strings.TrimSpace(req.Provider)
	if providerName == "" {
		providerName = cfg.LLMProvider
	}

	source := prompt.SourceContext{
		Title: req.SourceTitle, Excerpt: req.SourceExcerpt, Detail: req.SourceDetail, URL: req.HotURL,
	}
	router := h.app.Router()
	historySummary, promptHistory, _ := discussion.PrepareHistory(r.Context(), router, providerName, req.History)
	userContent := prompt.BuildUserContent(prompt.BuildOptions{
		Source: source, Question: req.Question,
		History: historyToPromptEntries(promptHistory), HistorySummary: historySummary,
		Round: round, GroupMode: false,
	})
	system, citations := h.app.AugmentSystemWithCitations(ch.ID, req.Question, req.SourceTitle, ch.BuildSystemMessage())
	llmReq := llm.ChatCompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: userContent},
		},
	}

	_ = sse.Event("meta", map[string]string{
		"conversationId": convID,
		"provider":       providerName,
	})

	resp, usedProvider, err := router.ChatStream(r.Context(), providerName, llmReq, func(chunk llm.StreamChunk) error {
		if chunk.Delta != "" {
			return sse.Event("delta", map[string]string{"content": chunk.Delta})
		}
		return nil
	})
	if err != nil {
		_ = sse.Event("error", map[string]string{"message": err.Error()})
		return
	}

	content := character.SanitizeReply(discussionExtract(resp))
	model := llmReq.Model
	if resp != nil && resp.Model != "" {
		model = resp.Model
	}
	if content == "" {
		_ = sse.Event("error", map[string]string{"message": "empty llm response"})
		return
	}

	store := h.app.Store
	if err := store.EnsureConversation(convID, "single", req.SourceTitle, req.HotURL, usedProvider, []string{req.CharacterID}); err != nil {
		_ = sse.Event("error", map[string]string{"message": err.Error()})
		return
	}
	if _, err := store.AddMessage(storage.Message{
		ConversationID: convID, Role: "user", Content: req.Question, Round: round,
	}); err != nil {
		_ = sse.Event("error", map[string]string{"message": err.Error()})
		return
	}
	if _, err := store.AddMessage(storage.Message{
		ConversationID: convID, Role: "assistant", CharacterID: ch.ID,
		CharacterName: ch.Name, Era: ch.Era, Content: content,
		Provider: usedProvider, Model: model, Round: round, Citations: citationsToStorage(citations),
	}); err != nil {
		_ = sse.Event("error", map[string]string{"message": err.Error()})
		return
	}
	_ = store.TouchConversation(convID)

	_ = sse.Event("done", map[string]any{
		"conversationId": convID,
		"content":        content,
		"provider":       usedProvider,
		"model":          model,
		"citations":      citationsToJSON(citations),
	})
}
