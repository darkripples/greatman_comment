package discussion

import (
	"context"
	"fmt"
	"strings"

	"renwen/server/internal/llm"
)

const summaryTrigger = 12
const summaryKeepRecent = 6

// PrepareHistory returns optional summary text and trimmed recent history for prompt building.
func PrepareHistory(ctx context.Context, router *llm.Router, provider string, history []HistoryItem) (summary string, recent []HistoryItem, err error) {
	if len(history) <= summaryTrigger {
		return "", history, nil
	}
	recent = history
	if len(recent) > summaryKeepRecent {
		recent = recent[len(recent)-summaryKeepRecent:]
	}
	summary, err = SummarizeHistory(ctx, router, provider, history[:len(history)-summaryKeepRecent])
	if err != nil {
		return "", history[len(history)-DefaultMaxHistory():], err
	}
	return summary, recent, nil
}

func DefaultMaxHistory() int { return 12 }

func SummarizeHistory(ctx context.Context, router *llm.Router, provider string, older []HistoryItem) (string, error) {
	if len(older) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("请将以下对话压缩为不超过300字的中文摘要，保留关键观点与分歧：\n\n")
	for _, h := range older {
		who := "用户"
		if h.Role == "assistant" && h.CharacterName != "" {
			who = h.CharacterName
		}
		b.WriteString(fmt.Sprintf("- %s：%s\n", who, h.Content))
	}
	resp, _, err := router.Chat(ctx, provider, llm.ChatCompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "你是讨论记录员，只输出摘要正文，不要标题或列表符号。"},
			{Role: "user", Content: b.String()},
		},
	})
	if err != nil {
		return "", err
	}
	content := extractContent(resp)
	if len([]rune(content)) > 350 {
		content = string([]rune(content)[:350]) + "…"
	}
	return strings.TrimSpace(content), nil
}
