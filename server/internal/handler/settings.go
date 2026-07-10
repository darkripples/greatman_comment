package handler

import (
	"encoding/json"
	"net/http"

	"renwen/server/internal/config"
	"renwen/server/internal/runtime"
)

type SettingsHandler struct {
	app *runtime.App
}

func NewSettingsHandler(app *runtime.App) *SettingsHandler {
	return &SettingsHandler{app: app}
}

type settingsResponse struct {
	config.AppSettings
	HasZhihuKey    bool   `json:"hasZhihuKey"`
	HasDeepSeekKey bool   `json:"hasDeepSeekKey"`
	Source         string `json:"source"`
}

type settingsPatch struct {
	config.AppSettings
	ZhihuMock *bool `json:"zhihuMock"`
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPut:
		h.put(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *SettingsHandler) get(w http.ResponseWriter, _ *http.Request) {
	app, err := h.app.Store.GetAppSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settingsResponse{
		AppSettings:    app,
		HasZhihuKey:    h.app.Keys.ZhihuAPIKey != "",
		HasDeepSeekKey: h.app.Keys.DeepSeekAPIKey != "",
		Source:         "sqlite",
	})
}

func (h *SettingsHandler) put(w http.ResponseWriter, r *http.Request) {
	current, err := h.app.Store.GetAppSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var patch settingsPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	next := mergeSettings(current, patch)
	if err := h.app.Store.SaveAppSettings(next); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.app.InvalidateCache()
	writeJSON(w, http.StatusOK, settingsResponse{
		AppSettings:    next,
		HasZhihuKey:    h.app.Keys.ZhihuAPIKey != "",
		HasDeepSeekKey: h.app.Keys.DeepSeekAPIKey != "",
		Source:         "sqlite",
	})
}

func mergeSettings(cur config.AppSettings, patch settingsPatch) config.AppSettings {
	p := patch.AppSettings
	if p.ApiEnvironment != "" {
		cur.ApiEnvironment = p.ApiEnvironment
	}
	if p.LocalApiMode != "" {
		cur.LocalApiMode = p.LocalApiMode
	}
	if p.ProdApiBase != "" {
		cur.ProdApiBase = p.ProdApiBase
	}
	if p.DevApiBase != "" {
		cur.DevApiBase = p.DevApiBase
	}
	if p.LLMProvider != "" {
		cur.LLMProvider = p.LLMProvider
	}
	if p.DeepSeekModel != "" {
		cur.DeepSeekModel = p.DeepSeekModel
	}
	if p.ZhidaModel != "" {
		cur.ZhidaModel = p.ZhidaModel
	}
	if p.DeepSeekAPIBase != "" {
		cur.DeepSeekAPIBase = p.DeepSeekAPIBase
	}
	if p.ZhihuAPIBase != "" {
		cur.ZhihuAPIBase = p.ZhihuAPIBase
	}
	if p.HotListCacheTTL != "" {
		cur.HotListCacheTTL = p.HotListCacheTTL
	}
	if p.HotListMinInterval != "" {
		cur.HotListMinInterval = p.HotListMinInterval
	}
	if p.SearchCacheTTL != "" {
		cur.SearchCacheTTL = p.SearchCacheTTL
	}
	if p.SearchMinInterval != "" {
		cur.SearchMinInterval = p.SearchMinInterval
	}
	if p.ChatMode != "" {
		cur.ChatMode = p.ChatMode
	}
	if patch.ZhihuMock != nil {
		cur.ZhihuMock = *patch.ZhihuMock
	}
	return cur
}
