package storage

import (
	"encoding/json"
	"log"
	"time"

	"renwen/server/internal/config"
)

const appSettingsKey = "app"

func (s *Store) ensureDefaultAppSettings() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM app_settings WHERE setting_key = ?`, appSettingsKey).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return s.saveAppSettingsRaw(config.DefaultAppSettings())
}

func (s *Store) GetAppSettings() (config.AppSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureDefaultAppSettings(); err != nil {
		return config.AppSettings{}, err
	}
	var raw string
	err := s.db.QueryRow(`SELECT setting_value FROM app_settings WHERE setting_key = ?`, appSettingsKey).Scan(&raw)
	if err != nil {
		return config.DefaultAppSettings(), err
	}
	var app config.AppSettings
	if err := json.Unmarshal([]byte(raw), &app); err != nil {
		return config.DefaultAppSettings(), err
	}
	return mergeAppDefaults(app), nil
}

func (s *Store) SaveAppSettings(app config.AppSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveAppSettingsRaw(mergeAppDefaults(app))
}

func (s *Store) saveAppSettingsRaw(app config.AppSettings) error {
	raw, err := json.Marshal(mergeAppDefaults(app))
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO app_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(setting_key) DO UPDATE SET setting_value = excluded.setting_value, updated_at = excluded.updated_at`,
		appSettingsKey, string(raw), time.Now().Unix(),
	)
	return err
}

// migrateAggressiveCacheIntervals bumps sub-hour min intervals that were likely
// misconfigured (e.g. README previously suggested matching the 5m TTL).
func (s *Store) migrateAggressiveCacheIntervals() error {
	app, err := s.GetAppSettings()
	if err != nil {
		return err
	}
	def := config.DefaultAppSettings()
	changed := false

	if d, err := time.ParseDuration(app.HotListMinInterval); err == nil && d < time.Hour {
		app.HotListMinInterval = def.HotListMinInterval
		changed = true
	}
	if d, err := time.ParseDuration(app.SearchMinInterval); err == nil && d < time.Hour {
		app.SearchMinInterval = def.HotListMinInterval
		changed = true
	}
	if !changed {
		return nil
	}
	log.Printf("[settings] migrated cache min intervals to hotList=%s search=%s", app.HotListMinInterval, app.SearchMinInterval)
	return s.SaveAppSettings(app)
}

func mergeAppDefaults(app config.AppSettings) config.AppSettings {
	def := config.DefaultAppSettings()
	if app.ApiEnvironment == "" {
		app.ApiEnvironment = def.ApiEnvironment
	}
	if app.LocalApiMode == "" {
		app.LocalApiMode = def.LocalApiMode
	}
	if app.ProdApiBase == "" {
		app.ProdApiBase = def.ProdApiBase
	}
	if app.DevApiBase == "" {
		app.DevApiBase = def.DevApiBase
	}
	if app.LLMProvider == "" {
		app.LLMProvider = def.LLMProvider
	}
	if app.DeepSeekModel == "" {
		app.DeepSeekModel = def.DeepSeekModel
	}
	if app.ZhidaModel == "" {
		app.ZhidaModel = def.ZhidaModel
	}
	if app.DeepSeekAPIBase == "" {
		app.DeepSeekAPIBase = def.DeepSeekAPIBase
	}
	if app.ZhihuAPIBase == "" {
		app.ZhihuAPIBase = def.ZhihuAPIBase
	}
	if app.HotListCacheTTL == "" {
		app.HotListCacheTTL = def.HotListCacheTTL
	}
	if app.HotListMinInterval == "" {
		app.HotListMinInterval = def.HotListMinInterval
	}
	if app.SearchCacheTTL == "" {
		app.SearchCacheTTL = def.SearchCacheTTL
	}
	if app.SearchMinInterval == "" {
		app.SearchMinInterval = def.SearchMinInterval
	}
	if app.ChatMode == "" {
		app.ChatMode = def.ChatMode
	}
	return app
}
