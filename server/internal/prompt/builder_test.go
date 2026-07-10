package prompt

import (
	"strings"
	"testing"
)

func TestBuildUserContent_includesExcerpt(t *testing.T) {
	out := BuildUserContent(BuildOptions{
		Source: SourceContext{
			Title:   "AI 会取代人类吗",
			Excerpt: "讨论人工智能对就业的影响",
		},
		Question: "你怎么看？",
		Round:    1,
	})
	if !containsAll(out, "【今日知乎议题】", "AI 会取代人类吗", "【议题摘要】", "就业的影响", "【用户提问】", "你怎么看？") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestBuildUserContent_truncatesHistory(t *testing.T) {
	history := make([]HistoryEntry, 15)
	for i := range history {
		history[i] = HistoryEntry{Role: "user", Content: "msg"}
	}
	out := BuildUserContent(BuildOptions{
		Question:   "继续",
		History:    history,
		MaxHistory: 12,
		Round:      2,
	})
	count := strings.Count(out, "- 用户：msg")
	if count != 12 {
		t.Fatalf("expected 12 history lines, got %d", count)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
