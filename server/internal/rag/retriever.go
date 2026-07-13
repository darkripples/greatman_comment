package rag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxPassageRunes = 200
const topK = 3

type Citation struct {
	Title   string `json:"title"`
	Source  string `json:"source,omitempty"`
	Excerpt string `json:"excerpt"`
}

type Passage struct {
	Title  string   `json:"title"`
	Text   string   `json:"text"`
	Tags   []string `json:"tags"`
	Source string   `json:"source,omitempty"`
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

func (r *Retriever) Augment(characterID, question, sourceTitle, base string) (string, []Citation) {
	if r == nil {
		return base, nil
	}
	passages := r.byCharacter[characterID]
	if len(passages) == 0 {
		return base, nil
	}
	query := strings.ToLower(question + " " + sourceTitle)
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return base, nil
	}

	type scored struct {
		p     Passage
		score int
	}
	scores := make([]scored, 0, len(passages))
	for _, p := range passages {
		score := 0
		for _, t := range tokens {
			if strings.Contains(strings.ToLower(p.Text), t) {
				score++
			}
			if strings.Contains(strings.ToLower(p.Title), t) {
				score += 3
			}
		}
		for _, tag := range p.Tags {
			tagLower := strings.ToLower(strings.TrimSpace(tag))
			if tagLower != "" && strings.Contains(query, tagLower) {
				score += 5
			}
		}
		if score > 0 {
			scores = append(scores, scored{p: p, score: score})
		}
	}
	if len(scores) == 0 {
		return base, nil
	}
	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].score != scores[j].score {
			return scores[i].score > scores[j].score
		}
		return scores[i].p.Title < scores[j].p.Title
	})
	selected := make([]scored, 0, topK)
	seenSources := make(map[string]bool)
	for _, candidate := range scores {
		source := strings.TrimSpace(candidate.p.Source)
		if source != "" && seenSources[source] {
			continue
		}
		if source != "" {
			seenSources[source] = true
		}
		selected = append(selected, candidate)
		if len(selected) == topK {
			break
		}
	}
	scores = selected

	citations := make([]Citation, 0, len(scores))
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
		citations = append(citations, Citation{
			Title:   s.p.Title,
			Source:  s.p.Source,
			Excerpt: text,
		})
	}
	return b.String(), citations
}

func (r *Retriever) AugmentSystemPrompt(characterID, question, sourceTitle, base string) string {
	prompt, _ := r.Augment(characterID, question, sourceTitle, base)
	return prompt
}

func tokenize(s string) []string {
	seen := make(map[string]bool)
	var tokens []string
	for _, part := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	}) {
		runes := []rune(strings.TrimSpace(part))
		if len(runes) < 2 {
			continue
		}
		addToken(&tokens, seen, string(runes))
		for i := 0; i+1 < len(runes); i++ {
			if isCJK(runes[i]) && isCJK(runes[i+1]) {
				addToken(&tokens, seen, string(runes[i:i+2]))
			}
		}
	}
	return tokens
}

func addToken(tokens *[]string, seen map[string]bool, token string) {
	if token != "" && !seen[token] {
		seen[token] = true
		*tokens = append(*tokens, token)
	}
}

func isCJK(r rune) bool { return r >= 0x4E00 && r <= 0x9FFF }

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "…"
}
