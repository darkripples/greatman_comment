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

type ChatHandler struct {
	app *runtime.App
}

type chatRequest struct {
	ConversationID string                   `json:"conversationId"`
	CharacterID    string                   `json:"characterId"`
	Question       string                   `json:"question"`
	SourceTitle    string                   `json:"sourceTitle"`
	SourceExcerpt  string                   `json:"sourceExcerpt"`
	SourceDetail   string                   `json:"sourceDetail"`
	HotURL         string                   `json:"hotUrl"`
	Provider       string                   `json:"provider"`
	Round          int                      `json:"round"`
	History        []discussion.HistoryItem `json:"history"`
}

type chatResponse struct {
	ConversationID string `json:"conversationId"`
	Content        string `json:"content"`
	Provider       string `json:"provider"`
	Model          string `json:"model"`
}

func NewChatHandler(app *runtime.App) *ChatHandler {
	return &ChatHandler{app: app}
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		Title:   req.SourceTitle,
		Excerpt: req.SourceExcerpt,
		Detail:  req.SourceDetail,
		URL:     req.HotURL,
	}

	router := h.app.Router()
	historySummary, promptHistory, _ := discussion.PrepareHistory(r.Context(), router, providerName, req.History)
	historyEntries := historyToPromptEntries(promptHistory)

	userContent := prompt.BuildUserContent(prompt.BuildOptions{
		Source:         source,
		Question:       req.Question,
		History:        historyEntries,
		HistorySummary: historySummary,
		Round:          round,
		GroupMode:      false,
	})

	system := h.app.AugmentSystem(ch.ID, req.Question, req.SourceTitle, ch.BuildSystemMessage())

	llmReq := llm.ChatCompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: userContent},
		},
	}

	resp, usedProvider, err := router.Chat(r.Context(), providerName, llmReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	content := character.SanitizeReply(discussionExtract(resp))
	model := llmReq.Model
	if resp != nil && resp.Model != "" {
		model = resp.Model
	}
	if content == "" {
		writeError(w, http.StatusBadGateway, "empty llm response")
		return
	}

	store := h.app.Store
	if err := store.EnsureConversation(convID, "single", req.SourceTitle, req.HotURL, usedProvider, []string{req.CharacterID}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := store.AddMessage(storage.Message{
		ConversationID: convID, Role: "user", Content: req.Question, Round: round,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := store.AddMessage(storage.Message{
		ConversationID: convID, Role: "assistant", CharacterID: ch.ID,
		CharacterName: ch.Name, Era: ch.Era, Content: content,
		Provider: usedProvider, Model: model, Round: round,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = store.TouchConversation(convID)

	writeJSON(w, http.StatusOK, chatResponse{
		ConversationID: convID,
		Content:        content,
		Provider:       usedProvider,
		Model:          model,
	})
}

func historyToPromptEntries(history []discussion.HistoryItem) []prompt.HistoryEntry {
	out := make([]prompt.HistoryEntry, 0, len(history))
	for _, h := range history {
		out = append(out, prompt.HistoryEntry{
			Role:          h.Role,
			CharacterName: h.CharacterName,
			Content:       h.Content,
		})
	}
	return out
}

func discussionExtract(resp *llm.ChatCompletionResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	return resp.Choices[0].Message.Content
}
