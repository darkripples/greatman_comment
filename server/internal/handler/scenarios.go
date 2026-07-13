package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type scenarioHotItem struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Excerpt    string `json:"excerpt,omitempty"`
	DetailText string `json:"detail_text,omitempty"`
}

type Scenario struct {
	ID             string          `json:"id"`
	Title          string          `json:"title"`
	Hook           string          `json:"hook"`
	Mode           string          `json:"mode"`
	CharacterIDs   []string        `json:"characterIds"`
	HotItem        scenarioHotItem `json:"hotItem"`
	SampleQuestion string          `json:"sampleQuestion"`
	ExpectedAngle  string          `json:"expectedAngle,omitempty"`
	DemoID         string          `json:"demoId,omitempty"`
}

type scenariosFile struct {
	Scenarios []Scenario `json:"scenarios"`
}

func loadScenarios() ([]Scenario, error) {
	for _, p := range []string{
		filepath.Join("config", "scenarios.json"),
		filepath.Join("server", "config", "scenarios.json"),
	} {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var f scenariosFile
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, err
		}
		return f.Scenarios, nil
	}
	return []Scenario{}, nil
}

func ScenariosHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, err := loadScenarios()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "load scenarios failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})
}
