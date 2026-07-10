package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const hotListSeedTTL = 30 * 24 * time.Hour

type hotListSeedItem struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Excerpt    string `json:"excerpt,omitempty"`
	DetailText string `json:"detail_text,omitempty"`
	Thumbnail  string `json:"thumbnail,omitempty"`
}

func (s *Store) seedHotListFixture() error {
	if _, err := readHotListFixtureFile(); err != nil {
		return nil
	}
	for _, limit := range []int{20} {
		key := fmt.Sprintf("hot_list:limit=%d", limit)
		var n int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM api_cache WHERE cache_key = ?`, key).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		if err := s.seedHotListKey(limit); err != nil {
			return err
		}
	}
	return nil
}

// EnsureHotListSeedForKey seeds backup hot list into api_cache when the key is missing.
func (s *Store) EnsureHotListSeedForKey(key string) error {
	const prefix = "hot_list:limit="
	if !strings.HasPrefix(key, prefix) {
		return nil
	}
	limit, err := strconv.Atoi(strings.TrimPrefix(key, prefix))
	if err != nil || limit <= 0 {
		return nil
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM api_cache WHERE cache_key = ?`, key).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return s.seedHotListKey(limit)
}

func (s *Store) seedHotListKey(limit int) error {
	raw, err := readHotListFixtureFile()
	if err != nil {
		return err
	}
	var items []hotListSeedItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("parse hot list fixture: %w", err)
	}
	if len(items) == 0 {
		return nil
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("hot_list:limit=%d", limit)
	return s.SetAPICache(key, string(payload), hotListSeedTTL)
}

func (s *Store) migrateHotListMockSeeds() error {
	for _, limit := range []int{20} {
		key := fmt.Sprintf("hot_list:limit=%d", limit)
		entry, ok := s.getAPICacheRow(key)
		if !ok {
			continue
		}
		var items []map[string]any
		if err := json.Unmarshal([]byte(entry.Payload), &items); err != nil {
			continue
		}
		if len(items) == 0 {
			continue
		}
		allMock := true
		for _, item := range items {
			if mock, ok := item["is_mock"].(bool); !ok || !mock {
				allMock = false
				break
			}
		}
		if !allMock {
			continue
		}
		if _, err := s.db.Exec(`DELETE FROM api_cache WHERE cache_key = ?`, key); err != nil {
			return err
		}
		if err := s.seedHotListKey(limit); err != nil {
			return err
		}
	}
	return nil
}

func readHotListFixtureFile() ([]byte, error) {
	for _, path := range []string{
		filepath.Join("fixtures", "hot_list.json"),
		filepath.Join("server", "fixtures", "hot_list.json"),
	} {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}
	return nil, os.ErrNotExist
}
