package discussion

import (
	"context"
	"fmt"
	"strings"

	"renwen/server/internal/llm"
	"renwen/server/internal/storage"
)

// GenerateRoundtableSummary produces a structured markdown summary for group discussions.
func GenerateRoundtableSummary(ctx context.Context, router *llm.Router, provider string, conv *storage.Conversation, msgs []storage.Message) (string, error) {
	if conv == nil || conv.Mode != "group" {
		return "", fmt.Errorf("roundtable summary requires a group conversation")
	}
	if len(msgs) == 0 {
		return "", fmt.Errorf("no messages to summarize")
	}

	var b strings.Builder
	b.WriteString("请根据以下群聊圆桌记录，输出结构化 Markdown 摘要，包含以下四级标题（必须全部出现）：\n")
	b.WriteString("## 议题\n## 各方核心观点（按人物）\n## 时代边界下的共识与分歧\n## 留给今人的一个问题\n\n")
	b.WriteString("要求：中文、简洁、每节 2-5 句；不要编造史料；保留各人物视角差异。\n\n")
	if conv.SourceTitle != "" {
		b.WriteString("【议题标题】")
		b.WriteString(conv.SourceTitle)
		b.WriteString("\n\n")
	}
	b.WriteString("【对话记录】\n")
	for _, m := range msgs {
		if m.Role == "user" {
			b.WriteString(fmt.Sprintf("用户：%s\n", m.Content))
			continue
		}
		name := m.CharacterName
		if name == "" {
			name = "assistant"
		}
		b.WriteString(fmt.Sprintf("%s：%s\n", name, m.Content))
	}

	resp, _, err := router.Chat(ctx, provider, llm.ChatCompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "你是人文讨论记录员，只输出 Markdown 正文，含指定标题，不要额外说明。"},
			{Role: "user", Content: b.String()},
		},
	})
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(extractContent(resp))
	if content == "" {
		return "", fmt.Errorf("empty summary response")
	}
	return content, nil
}
