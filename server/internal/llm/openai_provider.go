package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"renwen/server/internal/config"
)

type openAIProvider struct {
	name           string
	apiKey         string
	baseURL        string
	model          string
	completionPath string
	extraHeaders   func(*http.Request)
	client         *http.Client
}

func newOpenAIProvider(name, apiKey, baseURL, model, completionPath string, extraHeaders func(*http.Request)) *openAIProvider {
	return &openAIProvider{
		name:           name,
		apiKey:         apiKey,
		baseURL:        strings.TrimRight(baseURL, "/"),
		model:          model,
		completionPath: completionPath,
		extraHeaders:   extraHeaders,
		client:         &http.Client{Timeout: 120 * time.Second},
	}
}

func NewZhihuProvider(cfg config.Config) Provider {
	base := strings.TrimRight(cfg.ZhihuAPIBase, "/")
	return newOpenAIProvider("zhihu", cfg.ZhihuAPIKey, base, cfg.ZhidaModel, "/v1/chat/completions", func(req *http.Request) {
		req.Header.Set("X-Request-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "renwen/1.0")
	})
}

func NewDeepSeekProvider(cfg config.Config) Provider {
	return newOpenAIProvider("deepseek", cfg.DeepSeekAPIKey, cfg.DeepSeekAPIBase, cfg.DeepSeekModel, "/chat/completions", nil)
}

func (p *openAIProvider) Name() string  { return p.name }
func (p *openAIProvider) Model() string { return p.model }
func (p *openAIProvider) Available() bool {
	return strings.TrimSpace(p.apiKey) != ""
}

func (p *openAIProvider) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if !p.Available() {
		return nil, fmt.Errorf("%s api key not configured", p.name)
	}
	if req.Model == "" {
		req.Model = p.model
	}
	req.Stream = false

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		resp, err := p.doCompletion(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !retryable(err) {
			break
		}
	}
	return nil, lastErr
}

func (p *openAIProvider) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest, onChunk func(StreamChunk) error) (*ChatCompletionResponse, error) {
	if !p.Available() {
		return nil, fmt.Errorf("%s api key not configured", p.name)
	}
	if req.Model == "" {
		req.Model = p.model
	}
	req.Stream = true

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		resp, err := p.doStream(ctx, req, onChunk)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !retryable(err) {
			break
		}
	}
	return nil, lastErr
}

func (p *openAIProvider) doCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := p.newHTTPRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s chat completion failed (%d): %s", p.name, resp.StatusCode, string(raw))
	}
	var out ChatCompletionResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode %s response: %w", p.name, err)
	}
	return &out, nil
}

func (p *openAIProvider) doStream(ctx context.Context, req ChatCompletionRequest, onChunk func(StreamChunk) error) (*ChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := p.newHTTPRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s chat stream failed (%d): %s", p.name, resp.StatusCode, string(raw))
	}

	var fullContent strings.Builder
	var modelName string

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk streamPayload
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Model != "" {
			modelName = chunk.Model
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		fullContent.WriteString(delta)
		if onChunk != nil {
			if err := onChunk(StreamChunk{Delta: delta, Model: modelName}); err != nil {
				return nil, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	content := fullContent.String()
	if content == "" {
		return nil, fmt.Errorf("%s stream returned empty content", p.name)
	}
	if onChunk != nil {
		_ = onChunk(StreamChunk{Done: true, Model: modelName})
	}
	return &ChatCompletionResponse{
		Model: modelName,
		Choices: []ChatCompletionChoice{{
			Message: Message{Role: "assistant", Content: content},
		}},
	}, nil
}

func (p *openAIProvider) newHTTPRequest(ctx context.Context, body []byte) (*http.Request, error) {
	path := p.completionPath
	if path == "" {
		path = "/v1/chat/completions"
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	if p.extraHeaders != nil {
		p.extraHeaders(httpReq)
	}
	return httpReq, nil
}

type streamPayload struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func retryable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "(429)") ||
		strings.Contains(msg, "(502)") ||
		strings.Contains(msg, "(503)")
}
