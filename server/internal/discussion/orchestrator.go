package discussion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"renwen/server/internal/character"
	"renwen/server/internal/llm"
	"renwen/server/internal/prompt"
	"renwen/server/internal/rag"
)

const MaxRound = 5
const minGroupMembers = 2
const maxGroupMembers = 5

type HistoryItem struct {
	Role          string `json:"role"`
	CharacterID   string `json:"characterId,omitempty"`
	CharacterName string `json:"characterName,omitempty"`
	Content       string `json:"content"`
	Round         int    `json:"round,omitempty"`
}

type Turn struct {
	CharacterID string         `json:"characterId"`
	Name        string         `json:"name"`
	Era         string         `json:"era"`
	Content     string         `json:"content"`
	Round       int            `json:"round"`
	Citations   []rag.Citation `json:"citations,omitempty"`
}

type Orchestrator struct {
	characters   *character.Store
	llm          *llm.Router
	retriever    *rag.Retriever
	defaultLLM   string
	groupContext string
}

func NewOrchestrator(chars *character.Store, router *llm.Router, retriever *rag.Retriever, defaultLLM string) (*Orchestrator, error) {
	ctxText, err := loadGroupContext()
	if err != nil {
		ctxText = "你正在一场跨越时空的群聊中讨论今日议题。只输出你本人的发言，可回应他人。"
	}
	return &Orchestrator{
		characters:   chars,
		llm:          router,
		retriever:    retriever,
		defaultLLM:   defaultLLM,
		groupContext: ctxText,
	}, nil
}

func loadGroupContext() (string, error) {
	for _, p := range []string{
		filepath.Join("config", "group_context.txt"),
		filepath.Join("server", "config", "group_context.txt"),
	} {
		b, err := os.ReadFile(p)
		if err == nil {
			return strings.TrimSpace(string(b)), nil
		}
	}
	return "", fmt.Errorf("group_context.txt not found")
}

func (o *Orchestrator) Run(ctx context.Context, characterIDs []string, question string, source prompt.SourceContext, provider string, history []HistoryItem, round int) ([]Turn, string, string, error) {
	return o.runSpeaker(ctx, characterIDs, question, source, provider, history, round, -1, nil, nil)
}

// RunSpeaker runs one or all speakers. speakerIndex >= 0 runs only that index; priorInRoundTurns
// supplies earlier speakers in the same round when calling incrementally.
func (o *Orchestrator) RunSpeaker(ctx context.Context, characterIDs []string, question string, source prompt.SourceContext, provider string, history []HistoryItem, round, speakerIndex int, priorInRoundTurns []Turn) ([]Turn, string, string, error) {
	return o.runSpeaker(ctx, characterIDs, question, source, provider, history, round, speakerIndex, priorInRoundTurns, nil)
}

func (o *Orchestrator) RunSpeakerStream(ctx context.Context, characterIDs []string, question string, source prompt.SourceContext, provider string, history []HistoryItem, round, speakerIndex int, priorInRoundTurns []Turn, onDelta func(string) error) ([]Turn, string, string, error) {
	return o.runSpeaker(ctx, characterIDs, question, source, provider, history, round, speakerIndex, priorInRoundTurns, onDelta)
}

func (o *Orchestrator) runSpeaker(ctx context.Context, characterIDs []string, question string, source prompt.SourceContext, provider string, history []HistoryItem, round, speakerIndex int, priorInRoundTurns []Turn, onDelta func(string) error) ([]Turn, string, string, error) {
	if round <= 0 {
		round = 1
	}
	if round > MaxRound {
		return nil, "", "", fmt.Errorf("round exceeds max %d", MaxRound)
	}
	ids, err := validateCharacterIDs(characterIDs, o.characters)
	if err != nil {
		return nil, "", "", err
	}
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = o.defaultLLM
	}

	question = strings.TrimSpace(question)
	if question == "" {
		return nil, "", "", fmt.Errorf("question is required")
	}

	historySummary, promptHistory, _ := PrepareHistory(ctx, o.llm, provider, history)

	var turns []Turn
	priorInRound := make([]string, 0, len(priorInRoundTurns))
	for _, t := range priorInRoundTurns {
		priorInRound = append(priorInRound, fmt.Sprintf("%s：%s", t.Name, t.Content))
	}
	var usedModel string
	var usedProvider string

	start, end := 0, len(ids)
	if speakerIndex >= 0 {
		if speakerIndex >= len(ids) {
			return nil, "", "", fmt.Errorf("speakerIndex out of range")
		}
		start, end = speakerIndex, speakerIndex+1
	}

	historyEntries := historyToEntries(promptHistory)

	for i := start; i < end; i++ {
		cid := ids[i]
		ch, ok := o.characters.Get(cid)
		if !ok {
			continue
		}
		others := otherNames(ids, cid, o.characters)
		groupExtra := strings.ReplaceAll(o.groupContext, "{others}", others) + "\n\n" + roundInstruction(round)
		baseSystem := ch.BuildSystemMessage() + "\n\n" + groupExtra
		system := baseSystem
		var citations []rag.Citation
		if o.retriever != nil {
			system, citations = o.retriever.Augment(ch.ID, question, source.Title, baseSystem)
		}

		userContent := prompt.BuildUserContent(prompt.BuildOptions{
			Source:         source,
			Question:       question,
			History:        historyEntries,
			HistorySummary: historySummary,
			PriorInRound:   priorInRound,
			Round:          round,
			GroupMode:      true,
		})

		request := llm.ChatCompletionRequest{
			Messages: []llm.Message{
				{Role: "system", Content: system},
				{Role: "user", Content: userContent},
			},
		}
		var resp *llm.ChatCompletionResponse
		var prov string
		if onDelta == nil {
			resp, prov, err = o.llm.Chat(ctx, provider, request)
		} else {
			resp, prov, err = o.llm.ChatStream(ctx, provider, request, func(chunk llm.StreamChunk) error {
				if chunk.Delta == "" {
					return nil
				}
				return onDelta(chunk.Delta)
			})
		}
		usedProvider = prov
		if err != nil {
			return turns, usedProvider, usedModel, fmt.Errorf("%s 发言失败: %w", ch.Name, err)
		}

		content := character.SanitizeReply(extractContent(resp))
		if content == "" {
			return turns, usedProvider, usedModel, fmt.Errorf("empty response for %s", ch.Name)
		}
		if resp != nil && resp.Model != "" {
			usedModel = resp.Model
		}

		turns = append(turns, Turn{
			CharacterID: ch.ID,
			Name:        ch.Name,
			Era:         ch.Era,
			Content:     content,
			Round:       round,
			Citations:   citations,
		})
		priorInRound = append(priorInRound, fmt.Sprintf("%s：%s", ch.Name, content))
	}

	return turns, usedProvider, usedModel, nil
}

func roundInstruction(round int) string {
	if round <= 1 {
		return "第一轮请独立陈述立场，给出一个清晰判断，不必重复其他人的观点。"
	}
	return "本轮请明确回应至少一位先前发言者：指出认同或分歧，再补充你自己的判断。"
}

func historyToEntries(history []HistoryItem) []prompt.HistoryEntry {
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

func validateCharacterIDs(ids []string, store *character.Store) ([]string, error) {
	seen := make(map[string]bool)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		if _, ok := store.Get(id); !ok {
			return nil, fmt.Errorf("unknown characterId: %s", id)
		}
		seen[id] = true
		out = append(out, id)
	}
	if len(out) < minGroupMembers {
		return nil, fmt.Errorf("at least %d characterIds required", minGroupMembers)
	}
	if len(out) > maxGroupMembers {
		return nil, fmt.Errorf("at most %d characterIds allowed", maxGroupMembers)
	}
	return out, nil
}

func otherNames(ids []string, self string, store *character.Store) string {
	names := make([]string, 0, len(ids)-1)
	for _, id := range ids {
		if id == self {
			continue
		}
		if ch, ok := store.Get(id); ok {
			names = append(names, ch.Name)
		}
	}
	return strings.Join(names, "、")
}

func extractContent(resp *llm.ChatCompletionResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	return resp.Choices[0].Message.Content
}
