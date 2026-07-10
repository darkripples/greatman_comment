package runtime

import (
	"sync"

	"renwen/server/internal/character"
	"renwen/server/internal/config"
	"renwen/server/internal/discussion"
	"renwen/server/internal/llm"
	"renwen/server/internal/rag"
	"renwen/server/internal/storage"
	"renwen/server/internal/zhihu"
)

type App struct {
	Store      *storage.Store
	Characters *character.Store
	Keys       config.APIKeys
	retriever  *rag.Retriever

	mu            sync.Mutex
	cachedCfgKey  string
	cachedRouter  *llm.Router
	cachedOrch    *discussion.Orchestrator
}

func (a *App) Config() config.Config {
	app, err := a.Store.GetAppSettings()
	if err != nil {
		app = config.DefaultAppSettings()
	}
	cfg := config.Build(a.Keys, app)
	port, dataDir := config.Bootstrap()
	cfg.Port = port
	cfg.DataDir = dataDir
	return cfg
}

func (a *App) cfgCacheKey(cfg config.Config) string {
	return cfg.LLMProvider + "|" + cfg.DeepSeekModel + "|" + cfg.ZhidaModel + "|" +
		cfg.DeepSeekAPIBase + "|" + cfg.ZhihuAPIBase
}

func (a *App) Router() *llm.Router {
	cfg := a.Config()
	key := a.cfgCacheKey(cfg)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cachedRouter != nil && a.cachedCfgKey == key {
		return a.cachedRouter
	}
	a.cachedRouter = llm.NewRouter(cfg.LLMProvider,
		llm.NewDeepSeekProvider(cfg),
		llm.NewZhihuProvider(cfg),
	)
	a.cachedCfgKey = key
	a.cachedOrch = nil
	return a.cachedRouter
}

func (a *App) Orchestrator() (*discussion.Orchestrator, error) {
	cfg := a.Config()
	key := a.cfgCacheKey(cfg)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cachedOrch != nil && a.cachedCfgKey == key {
		return a.cachedOrch, nil
	}
	router := llm.NewRouter(cfg.LLMProvider,
		llm.NewDeepSeekProvider(cfg),
		llm.NewZhihuProvider(cfg),
	)
	orch, err := discussion.NewOrchestrator(a.Characters, router, a.Retriever(), cfg.LLMProvider)
	if err != nil {
		return nil, err
	}
	a.cachedRouter = router
	a.cachedOrch = orch
	a.cachedCfgKey = key
	return orch, nil
}

func (a *App) InvalidateCache() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cachedCfgKey = ""
	a.cachedRouter = nil
	a.cachedOrch = nil
}

func (a *App) ZhihuClient() *zhihu.Client {
	return zhihu.NewClient(a.Config())
}

func (a *App) Retriever() *rag.Retriever {
	if a.retriever != nil {
		return a.retriever
	}
	r, err := rag.NewRetriever()
	if err != nil {
		return &rag.Retriever{}
	}
	a.retriever = r
	return a.retriever
}

func (a *App) AugmentSystem(characterID, question, sourceTitle, base string) string {
	return a.Retriever().AugmentSystemPrompt(characterID, question, sourceTitle, base)
}
