package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadAPIKeys() APIKeys {
	zhihuKey := strings.TrimSpace(os.Getenv("ZHIHU_API_KEY"))
	if zhihuKey == "" {
		zhihuKey = strings.TrimSpace(os.Getenv("ZHIHU_ACCESS_SECRET"))
	}
	keys := APIKeys{
		ZhihuAPIKey:    zhihuKey,
		DeepSeekAPIKey: strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
	}
	if keys.ZhihuAPIKey == "" {
		log.Println("[warn] ZHIHU_API_KEY missing: hot-list/search/zhihu LLM need env key")
	}
	if keys.DeepSeekAPIKey == "" {
		log.Println("[warn] DEEPSEEK_API_KEY missing: deepseek LLM need env key")
	}
	return keys
}

func Bootstrap() (port, dataDir string) {
	port = "30302"
	if v := strings.TrimSpace(os.Getenv("SERVER_PORT")); v != "" {
		port = v
	} else if v := strings.TrimSpace(os.Getenv("PORT")); v != "" {
		port = v
	}
	dataDir = "./data"
	if v := strings.TrimSpace(os.Getenv("RENWEN_DATA_DIR")); v != "" {
		dataDir = v
	}
	return port, dataDir
}

func parseDuration(raw string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return d
}

func ParseLimit(raw string, defaultVal, max int) int {
	if raw == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultVal
	}
	if n > max {
		return max
	}
	return n
}
