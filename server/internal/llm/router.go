package llm

import (
	"context"
	"fmt"
	"strings"
)

type Router struct {
	providers       map[string]Provider
	defaultProvider string
}

func NewRouter(defaultProvider string, providers ...Provider) *Router {
	m := make(map[string]Provider, len(providers))
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &Router{providers: m, defaultProvider: defaultProvider}
}

func (r *Router) Resolve(name string) (Provider, error) {
	id := strings.TrimSpace(name)
	if id == "" {
		id = r.defaultProvider
	}
	p, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("unknown llm provider: %s", id)
	}
	if !p.Available() {
		return nil, fmt.Errorf("llm provider unavailable: %s", id)
	}
	return p, nil
}

func (r *Router) List() []ProviderInfo {
	order := []string{"deepseek", "zhihu"}
	seen := make(map[string]bool)
	out := make([]ProviderInfo, 0, len(r.providers))

	appendInfo := func(id string) {
		p, ok := r.providers[id]
		if !ok || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, ProviderInfo{
			ID:        p.Name(),
			Name:      displayName(p.Name()),
			Available: p.Available(),
			Model:     p.Model(),
			Default:   id == r.defaultProvider,
		})
	}

	for _, id := range order {
		appendInfo(id)
	}
	for id := range r.providers {
		appendInfo(id)
	}
	return out
}

func (r *Router) Chat(ctx context.Context, providerName string, req ChatCompletionRequest) (*ChatCompletionResponse, string, error) {
	p, err := r.Resolve(providerName)
	if err != nil {
		return nil, "", err
	}
	if req.Model == "" {
		req.Model = p.Model()
	}
	resp, err := p.ChatCompletion(ctx, req)
	if err != nil {
		return nil, p.Name(), err
	}
	return resp, p.Name(), nil
}

func (r *Router) ChatStream(ctx context.Context, providerName string, req ChatCompletionRequest, onChunk func(StreamChunk) error) (*ChatCompletionResponse, string, error) {
	p, err := r.Resolve(providerName)
	if err != nil {
		return nil, "", err
	}
	if req.Model == "" {
		req.Model = p.Model()
	}
	resp, err := p.ChatCompletionStream(ctx, req, onChunk)
	if err != nil {
		// fallback to blocking completion
		resp, fbErr := p.ChatCompletion(ctx, req)
		if fbErr != nil {
			return nil, p.Name(), err
		}
		content := ""
		if resp != nil && len(resp.Choices) > 0 {
			content = resp.Choices[0].Message.Content
		}
		if onChunk != nil && content != "" {
			_ = onChunk(StreamChunk{Delta: content, Model: resp.Model})
			_ = onChunk(StreamChunk{Done: true, Model: resp.Model})
		}
		return resp, p.Name(), nil
	}
	return resp, p.Name(), nil
}

func displayName(id string) string {
	switch id {
	case "zhihu":
		return "知乎直答"
	case "deepseek":
		return "DeepSeek"
	default:
		return id
	}
}
