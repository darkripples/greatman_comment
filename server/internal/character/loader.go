package character

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Character struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Era          string   `json:"era"`
	Summary      string   `json:"summary"`
	SystemPrompt string   `json:"system_prompt"`
	CitationHint string   `json:"citation_hint"`
	Portrait     string   `json:"portrait,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	KeyWorks     []string `json:"keyWorks,omitempty"`
	FitTopics    []string `json:"fitTopics,omitempty"`
	Intro        string   `json:"intro,omitempty"`
}

type PublicCharacter struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Era       string   `json:"era"`
	Summary   string   `json:"summary"`
	Portrait  string   `json:"portrait,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	KeyWorks  []string `json:"keyWorks,omitempty"`
	FitTopics []string `json:"fitTopics,omitempty"`
	Intro     string   `json:"intro,omitempty"`
}

type file struct {
	Characters []Character `json:"characters"`
}

type Store struct {
	byID map[string]Character
	list []PublicCharacter
}

func LoadFromFile(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read characters: %w", err)
	}
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse characters: %w", err)
	}
	store := &Store{byID: make(map[string]Character), list: make([]PublicCharacter, 0, len(f.Characters))}
	for _, c := range f.Characters {
		store.byID[c.ID] = c
		store.list = append(store.list, PublicCharacter{
			ID: c.ID, Name: c.Name, Era: c.Era, Summary: c.Summary,
			Portrait: c.Portrait, Tags: c.Tags, KeyWorks: c.KeyWorks,
			FitTopics: c.FitTopics, Intro: c.Intro,
		})
	}
	return store, nil
}

func DefaultPath() string {
	return filepath.Join("config", "characters.json")
}

func (s *Store) List() []PublicCharacter { return s.list }

func (s *Store) Get(id string) (Character, bool) {
	c, ok := s.byID[id]
	return c, ok
}

func (c Character) BuildSystemMessage() string {
	return c.SystemPrompt + "\n\n" + c.CitationHint
}
