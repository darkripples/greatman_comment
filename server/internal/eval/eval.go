package eval

import (
	"context"
	"strings"
	"time"

	"renwen/server/internal/character"
	"renwen/server/internal/llm"
	"renwen/server/internal/rag"
)

type Question struct{ Category, Prompt string }
type Finding struct{ Severity, Rule, Detail string }
type Result struct {
	CharacterID, CharacterName, Category, Question, Reply, Provider, Model string
	Citations                                                              []rag.Citation `json:"citations,omitempty"`
	Findings                                                               []Finding      `json:"findings,omitempty"`
}
type Report struct {
	GeneratedAt                             time.Time `json:"generatedAt"`
	Provider                                string    `json:"provider"`
	Results                                 []Result  `json:"results"`
	SevereCount, PassedCount, CitationCount int
}

func DefaultQuestions() []Question {
	return []Question{
		{"modern", "你怎么看今天的 AI 写作？"}, {"era", "请具体说说 2020 年发生了什么？"}, {"stance", "面对困境，人应该怎样行动？"}, {"slang", "请用网络梗评价内卷。"}, {"evidence", "你的核心思想为何值得今天的人参考？"},
	}
}

func Run(ctx context.Context, characters *character.Store, router *llm.Router, retriever *rag.Retriever, provider string, ids []string) (Report, error) {
	report := Report{GeneratedAt: time.Now().UTC(), Provider: provider}
	for _, id := range ids {
		ch, ok := characters.Get(id)
		if !ok {
			continue
		}
		for _, q := range DefaultQuestions() {
			system, citations := retriever.Augment(ch.ID, q.Prompt, "", ch.BuildSystemMessage())
			resp, used, err := router.Chat(ctx, provider, llm.ChatCompletionRequest{Messages: []llm.Message{{Role: "system", Content: system}, {Role: "user", Content: q.Prompt}}})
			if err != nil {
				return report, err
			}
			reply, model := "", ""
			if resp != nil {
				model = resp.Model
				if len(resp.Choices) > 0 {
					reply = resp.Choices[0].Message.Content
				}
			}
			result := Result{CharacterID: id, CharacterName: ch.Name, Category: q.Category, Question: q.Prompt, Reply: reply, Provider: used, Model: model, Citations: citations}
			result.Findings = inspect(q.Category, reply, citations)
			for _, finding := range result.Findings {
				if finding.Severity == "severe" {
					report.SevereCount++
				}
			}
			if len(citations) > 0 {
				report.CitationCount++
			}
			if len(result.Findings) == 0 {
				report.PassedCount++
			}
			report.Results = append(report.Results, result)
		}
	}
	return report, nil
}

func inspect(category, reply string, citations []rag.Citation) []Finding {
	var findings []Finding
	lower := strings.ToLower(reply)
	if category == "era" && !containsAny(reply, "不知", "未闻", "局限", "不可确知") {
		findings = append(findings, Finding{"severe", "era-boundary", "超时代问题未说明时代局限"})
	}
	if containsAny(lower, "yyds", "666", "emo", "绝绝子") {
		findings = append(findings, Finding{"warning", "modern-slang", "回复包含明显网络表达"})
	}
	if category == "evidence" && len(citations) == 0 {
		findings = append(findings, Finding{"warning", "missing-citation", "史料相关问题没有命中引用"})
	}
	return findings
}

func containsAny(text string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(text, value) {
			return true
		}
	}
	return false
}
