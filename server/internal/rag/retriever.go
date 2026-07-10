package rag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const maxPassageRunes = 200
const topK = 3

type Passage struct {
	Title string   `json:"title"`
	Text  string   `json:"text"`
	Tags  []string `json:"tags"`
}

type sourceFile struct {
	CharacterID string    `json:"characterId"`
	Passages    []Passage `json:"passages"`
}

type Retriever struct {
	byCharacter map[string][]Passage
}

func NewRetriever() (*Retriever, error) {
	r := &Retriever{byCharacter: make(map[string][]Passage)}
	for _, dir := range []string{
		filepath.Join("config", "sources"),
		filepath.Join("server", "config", "sources"),
	} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var f sourceFile
			if err := json.Unmarshal(raw, &f); err != nil || f.CharacterID == "" {
				continue
			}
			r.byCharacter[f.CharacterID] = append(r.byCharacter[f.CharacterID], f.Passages...)
		}
		return r, nil
	}
	return r, nil
}

func (r *Retriever) AugmentSystemPrompt(characterID, question, sourceTitle, base string) string {
	if r == nil {
		return base
	}
	passages := r.byCharacter[characterID]
	if len(passages) == 0 {
		return base
	}
	query := strings.ToLower(question + " " + sourceTitle)
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return base
	}

	type scored struct {
		p     Passage
		score int
	}
	scores := make([]scored, 0, len(passages))
	for _, p := range passages {
		text := strings.ToLower(p.Title + " " + p.Text + " " + strings.Join(p.Tags, " "))
		score := 0
		for _, t := range tokens {
			if strings.Contains(text, t) {
				score++
			}
		}
		for _, tag := range p.Tags {
			tagLower := strings.ToLower(strings.TrimSpace(tag))
			if tagLower != "" && strings.Contains(query, tagLower) {
				score += 2
			}
		}
		if score > 0 {
			scores = append(scores, scored{p: p, score: score})
		}
	}
	if len(scores) == 0 {
		return base
	}
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	if len(scores) > topK {
		scores = scores[:topK]
	}

	var b strings.Builder
	b.WriteString(base)
	b.WriteString("\n\n【可参考史料片段，勿捏造未列出内容】\n")
	for _, s := range scores {
		text := truncateRunes(strings.TrimSpace(s.p.Text), maxPassageRunes)
		if s.p.Title != "" {
			b.WriteString(fmt.Sprintf("- %s：%s\n", s.p.Title, text))
		} else {
			b.WriteString("- ")
			b.WriteString(text)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func tokenize(s string) []string {
	s = strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, s)
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '，' || r == '。' || r == '？' || r == '！' || r == '、' || r == ',' || r == '.' || r == '?' || r == '!'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if utf8.RuneCountInString(p) >= 2 {
			out = append(out, p)
		}
	}
	return out
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "…"
}
