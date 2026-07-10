package prompt

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const MaxExcerptRunes = 500
const MaxDetailRunes = 800
const DefaultMaxHistory = 12

type SourceContext struct {
	Title   string
	Excerpt string
	Detail  string
	URL     string
}

type HistoryEntry struct {
	Role          string
	CharacterName string
	Content       string
}

type BuildOptions struct {
	Source         SourceContext
	Question       string
	History        []HistoryEntry
	HistorySummary string
	PriorInRound   []string
	Round          int
	MaxHistory     int
	GroupMode      bool
}

func BuildUserContent(opts BuildOptions) string {
	if opts.MaxHistory <= 0 {
		opts.MaxHistory = DefaultMaxHistory
	}
	if opts.Round <= 0 {
		opts.Round = 1
	}

	var b strings.Builder
	writeSourceSection(&b, opts.Source)

	if summary := strings.TrimSpace(opts.HistorySummary); summary != "" {
		b.WriteString("【此前讨论摘要】\n")
		b.WriteString(summary)
		b.WriteString("\n\n")
	}

	history := opts.History
	if len(history) > opts.MaxHistory {
		history = history[len(history)-opts.MaxHistory:]
	}
	if len(history) > 0 {
		b.WriteString("【此前讨论记录】\n")
		for _, h := range history {
			who := "用户"
			if h.Role == "assistant" && h.CharacterName != "" {
				who = h.CharacterName
			}
			b.WriteString(fmt.Sprintf("- %s：%s\n", who, h.Content))
		}
		b.WriteString("\n")
	}

	if opts.GroupMode {
		b.WriteString(fmt.Sprintf("【当前为第 %d 轮发言】\n", opts.Round))
		b.WriteString("【本轮用户发言或追问】")
	} else {
		if opts.Round > 1 {
			b.WriteString(fmt.Sprintf("【当前为第 %d 轮对话】\n", opts.Round))
		}
		b.WriteString("【用户提问】")
	}
	b.WriteString(strings.TrimSpace(opts.Question))

	if len(opts.PriorInRound) > 0 {
		b.WriteString("\n\n【本轮已发言】\n")
		for _, p := range opts.PriorInRound {
			b.WriteString("- ")
			b.WriteString(p)
			b.WriteString("\n")
		}
		b.WriteString("\n请你在上述语境下发表你的看法。")
	}

	return b.String()
}

func writeSourceSection(b *strings.Builder, source SourceContext) {
	title := strings.TrimSpace(source.Title)
	if title == "" {
		return
	}
	b.WriteString("【今日知乎议题】")
	b.WriteString(title)
	b.WriteString("\n")
	if ex := truncateRunes(strings.TrimSpace(source.Excerpt), MaxExcerptRunes); ex != "" {
		b.WriteString("【议题摘要】")
		b.WriteString(ex)
		b.WriteString("\n")
	}
	if det := truncateRunes(strings.TrimSpace(source.Detail), MaxDetailRunes); det != "" {
		b.WriteString("【议题详情】")
		b.WriteString(det)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func truncateRunes(s string, max int) string {
	if s == "" || max <= 0 {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}

func ToHistoryEntries[T any](items []T, fn func(T) HistoryEntry) []HistoryEntry {
	out := make([]HistoryEntry, 0, len(items))
	for _, item := range items {
		out = append(out, fn(item))
	}
	return out
}
