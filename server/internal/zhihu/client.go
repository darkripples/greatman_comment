package zhihu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"renwen/server/internal/config"
)

type HotItem struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Excerpt    string `json:"excerpt,omitempty"`
	DetailText string `json:"detail_text,omitempty"`
	Thumbnail  string `json:"thumbnail,omitempty"`
	IsMock     bool   `json:"is_mock,omitempty"`
}

type SearchItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Excerpt string `json:"excerpt,omitempty"`
	Type    string `json:"type,omitempty"`
	IsMock  bool   `json:"is_mock,omitempty"`
}

type Client struct {
	apiKey  string
	baseURL string
	mock    bool
	client  *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		apiKey:  cfg.ZhihuAPIKey,
		baseURL: strings.TrimRight(cfg.ZhihuAPIBase, "/"),
		mock:    cfg.ZhihuMock,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) HotList(ctx context.Context, limit int) ([]HotItem, error) {
	if c.mock {
		return loadFixtureHotList(limit)
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("ZHIHU_API_KEY not configured")
	}
	raw, err := c.get(ctx, "/api/v1/content/hot_list", url.Values{
		"Limit": {strconv.Itoa(limit)},
	})
	if err != nil {
		return nil, err
	}
	items, err := parseHotList(raw)
	if err != nil {
		return nil, err
	}
	return ensureHotItems(items), nil
}

func (c *Client) Search(ctx context.Context, query string, count int) ([]SearchItem, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if c.mock {
		return mockSearch(query, count), nil
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("ZHIHU_API_KEY not configured")
	}
	raw, err := c.get(ctx, "/api/v1/content/zhihu_search", url.Values{
		"Query": {query},
		"Count": {strconv.Itoa(count)},
	})
	if err != nil {
		return nil, err
	}
	items, err := parseSearch(raw)
	if err != nil {
		return nil, err
	}
	return ensureSearchItems(items), nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "renwen/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("zhihu api failed (%d): %s", resp.StatusCode, string(body))
	}
	if err := checkAPIEnvelope(body); err != nil {
		return nil, err
	}
	return body, nil
}

type apiEnvelope struct {
	Code    int             `json:"Code"`
	Message string          `json:"Message"`
	Data    json.RawMessage `json:"Data"`
}

func checkAPIEnvelope(raw []byte) error {
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil
	}
	if env.Code == 0 {
		return nil
	}
	if env.Message != "" {
		return fmt.Errorf("zhihu api error (code %d): %s", env.Code, env.Message)
	}
	return fmt.Errorf("zhihu api error (code %d)", env.Code)
}

func loadFixtureHotList(limit int) ([]HotItem, error) {
	for _, path := range []string{
		filepath.Join("fixtures", "hot_list.json"),
		filepath.Join("server", "fixtures", "hot_list.json"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var items []HotItem
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, err
		}
		for i := range items {
			items[i].IsMock = true
		}
		if limit > 0 && len(items) > limit {
			items = items[:limit]
		}
		return ensureHotItems(items), nil
	}
	return nil, fmt.Errorf("load fixture hot list: file not found")
}

func mockSearch(query string, count int) []SearchItem {
	items := []SearchItem{
		{Title: "关于「" + query + "」的知乎讨论", URL: "https://www.zhihu.com/", Excerpt: "mock 搜索结果", Type: "answer", IsMock: true},
		{Title: query + " 相关历史视角", URL: "https://www.zhihu.com/", Excerpt: "mock 补充语境", Type: "article", IsMock: true},
	}
	if count > 0 && len(items) > count {
		items = items[:count]
	}
	return items
}

func parseHotList(raw []byte) ([]HotItem, error) {
	if items, ok := parseOfficialHotItems(raw); ok {
		return items, nil
	}

	var direct []HotItem
	if err := json.Unmarshal(raw, &direct); err == nil && len(direct) > 0 {
		return direct, nil
	}
	return []HotItem{}, fmt.Errorf("unexpected hot list payload: %s", truncate(string(raw), 300))
}

func parseOfficialHotItems(raw []byte) ([]HotItem, bool) {
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil, false
	}

	type apiHotItem struct {
		Title        string `json:"Title"`
		URL          string `json:"Url"`
		Summary      string `json:"Summary"`
		ThumbnailURL string `json:"ThumbnailUrl"`
	}
	type dataBlock struct {
		Items []apiHotItem `json:"Items"`
	}

	var block dataBlock
	if err := json.Unmarshal(env.Data, &block); err != nil || len(block.Items) == 0 {
		return nil, false
	}

	out := make([]HotItem, 0, len(block.Items))
	for _, item := range block.Items {
		if strings.TrimSpace(item.Title) == "" {
			continue
		}
		out = append(out, HotItem{
			Title:     item.Title,
			URL:       item.URL,
			Excerpt:   item.Summary,
			Thumbnail: item.ThumbnailURL,
		})
	}
	return out, len(out) > 0
}

func parseSearch(raw []byte) ([]SearchItem, error) {
	if items, ok := parseOfficialSearchItems(raw); ok {
		return items, nil
	}

	var direct []SearchItem
	if err := json.Unmarshal(raw, &direct); err == nil && len(direct) > 0 {
		return direct, nil
	}
	return []SearchItem{}, fmt.Errorf("unexpected search payload: %s", truncate(string(raw), 300))
}

func parseOfficialSearchItems(raw []byte) ([]SearchItem, bool) {
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil, false
	}

	type apiSearchItem struct {
		Title       string `json:"Title"`
		URL         string `json:"Url"`
		ContentText string `json:"ContentText"`
	}
	type dataBlock struct {
		Items []apiSearchItem `json:"Items"`
	}

	var block dataBlock
	if err := json.Unmarshal(env.Data, &block); err != nil || len(block.Items) == 0 {
		return nil, false
	}

	out := make([]SearchItem, 0, len(block.Items))
	for _, item := range block.Items {
		if strings.TrimSpace(item.Title) == "" {
			continue
		}
		out = append(out, SearchItem{
			Title:   item.Title,
			URL:     item.URL,
			Excerpt: item.ContentText,
		})
	}
	return out, len(out) > 0
}

func ensureHotItems(items []HotItem) []HotItem {
	if items == nil {
		return []HotItem{}
	}
	return items
}

func ensureSearchItems(items []SearchItem) []SearchItem {
	if items == nil {
		return []SearchItem{}
	}
	return items
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
