package config

import "time"

// APIKeys 仅来自环境变量（敏感信息）。
type APIKeys struct {
	ZhihuAPIKey    string
	DeepSeekAPIKey string
}

// AppSettings 来自 SQLite，由设置页读写。
type AppSettings struct {
	ApiEnvironment     string `json:"apiEnvironment"`
	LocalApiMode       string `json:"localApiMode"`
	ProdApiBase        string `json:"prodApiBase"`
	DevApiBase         string `json:"devApiBase"`
	LLMProvider        string `json:"llmProvider"`
	DeepSeekModel      string `json:"deepseekModel"`
	ZhidaModel         string `json:"zhidaModel"`
	DeepSeekAPIBase    string `json:"deepseekApiBase"`
	ZhihuAPIBase       string `json:"zhihuApiBase"`
	ZhihuMock          bool   `json:"zhihuMock"`
	HotListCacheTTL    string `json:"hotListCacheTtl"`
	HotListMinInterval string `json:"hotListMinInterval"`
	SearchCacheTTL     string `json:"searchCacheTtl"`
	SearchMinInterval  string `json:"searchMinInterval"`
	ChatMode           string `json:"chatMode"`
}

func DefaultAppSettings() AppSettings {
	return AppSettings{
		ApiEnvironment:     "local",
		LocalApiMode:       "rewrite",
		ProdApiBase:        "https://your-prod-api.example.com",
		DevApiBase:         "http://127.0.0.1:30302",
		LLMProvider:        "deepseek",
		DeepSeekModel:      "deepseek-v4-flash",
		ZhidaModel:         "zhida-fast-1p5",
		DeepSeekAPIBase:    "https://api.deepseek.com",
		ZhihuAPIBase:       "https://developer.zhihu.com",
		ZhihuMock:          false,
		HotListCacheTTL:    "5m",
		HotListMinInterval: "5h",
		SearchCacheTTL:     "5m",
		SearchMinInterval:  "5m",
		ChatMode:           "single",
	}
}

// Config 运行时合并配置：Key 来自环境变量，其余来自 AppSettings。
type Config struct {
	Port               string
	DataDir            string
	ZhihuAPIKey        string
	ZhihuAPIBase       string
	ZhihuMock          bool
	DeepSeekAPIKey     string
	DeepSeekAPIBase    string
	DeepSeekModel      string
	ZhidaModel         string
	LLMProvider        string
	HotListCacheTTL    time.Duration
	HotListMinInterval time.Duration
	SearchCacheTTL     time.Duration
	SearchMinInterval  time.Duration
	App                AppSettings
}

func Build(keys APIKeys, app AppSettings) Config {
	return Config{
		Port:               "30302",
		DataDir:            "./data",
		ZhihuAPIKey:        keys.ZhihuAPIKey,
		ZhihuAPIBase:       app.ZhihuAPIBase,
		ZhihuMock:          app.ZhihuMock,
		DeepSeekAPIKey:     keys.DeepSeekAPIKey,
		DeepSeekAPIBase:    app.DeepSeekAPIBase,
		DeepSeekModel:      app.DeepSeekModel,
		ZhidaModel:         app.ZhidaModel,
		LLMProvider:        app.LLMProvider,
		HotListCacheTTL:    parseDuration(app.HotListCacheTTL, 5*time.Minute),
		HotListMinInterval: parseDuration(app.HotListMinInterval, 5*time.Hour),
		SearchCacheTTL:     parseDuration(app.SearchCacheTTL, 5*time.Minute),
		SearchMinInterval:  parseDuration(app.SearchMinInterval, 5*time.Minute),
		App:                app,
	}
}
