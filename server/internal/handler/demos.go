package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type DemoMessage struct {
	Role          string `json:"role"`
	CharacterID   string `json:"characterId,omitempty"`
	CharacterName string `json:"characterName,omitempty"`
	Era           string `json:"era,omitempty"`
	Content       string `json:"content"`
	Round         int    `json:"round,omitempty"`
}

type DemoConversation struct {
	ID           string        `json:"id"`
	ScenarioID   string        `json:"scenarioId,omitempty"`
	SourceTitle  string        `json:"sourceTitle"`
	Mode         string        `json:"mode"`
	CharacterIDs []string      `json:"characterIds,omitempty"`
	Messages     []DemoMessage `json:"messages"`
}

type demosFile struct {
	Demos []DemoConversation `json:"demos"`
}

func loadDemos() ([]DemoConversation, error) {
	for _, p := range []string{
		filepath.Join("fixtures", "demo_conversations.json"),
		filepath.Join("server", "fixtures", "demo_conversations.json"),
	} {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var f demosFile
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, err
		}
		return f.Demos, nil
	}
	return []DemoConversation{}, nil
}

func DemosHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := loadDemos()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "load demos failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})
}
