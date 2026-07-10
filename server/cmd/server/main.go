package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"renwen/server/internal/character"
	"renwen/server/internal/config"
	"renwen/server/internal/handler"
	"renwen/server/internal/hotlist"
	"renwen/server/internal/runtime"
	"renwen/server/internal/storage"
)

func main() {
	keys := config.LoadAPIKeys()
	port, dataDir := config.Bootstrap()
	charPath := resolvePath(character.DefaultPath())

	characters, err := character.LoadFromFile(charPath)
	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.Open(resolveDataDir(dataDir))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	app := &runtime.App{Store: store, Characters: characters, Keys: keys}
	cfg := app.Config()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hotlist.StartRefresher(ctx, app)

	mux := http.NewServeMux()
	mux.Handle("GET /api/health", withCORS(http.HandlerFunc(handler.Health)))
	mux.Handle("GET /api/settings", withCORS(handler.NewSettingsHandler(app)))
	mux.Handle("PUT /api/settings", withCORS(handler.NewSettingsHandler(app)))
	mux.Handle("GET /api/hot-list", withCORS(handler.NewHotListHandler(app)))
	mux.Handle("GET /api/search", withCORS(handler.NewSearchHandler(app)))
	mux.Handle("GET /api/characters", withCORS(handler.NewCharactersHandler(characters)))
	mux.Handle("GET /api/providers", withCORS(handler.NewProvidersHandler(app)))
	mux.Handle("POST /api/chat", withCORS(handler.NewChatHandler(app)))
	mux.Handle("POST /api/chat/stream", withCORS(handler.NewChatStreamHandler(app)))
	mux.Handle("POST /api/group-discuss", withCORS(handler.NewGroupDiscussHandler(app)))
	mux.Handle("POST /api/group-discuss/stream", withCORS(handler.NewGroupDiscussStreamHandler(app)))
	mux.Handle("/api/conversations", withCORS(handler.NewConversationsHandler(store)))
	mux.Handle("/api/conversations/", withCORS(handler.NewConversationsHandler(store)))

	addr := ":" + port
	log.Printf("renwen server listening on %s (db=%s/renwen.db, settings=sqlite, llm=%s, zhihu_mock=%v)",
		addr, resolveDataDir(dataDir), cfg.LLMProvider, cfg.ZhihuMock)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func resolvePath(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	alt := filepath.Join("server", path)
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return path
}

func resolveDataDir(dir string) string {
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	alt := filepath.Join("server", dir)
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return dir
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if originAllowed(origin, defaultCORSOrigins()) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func defaultCORSOrigins() []string {
	return []string{
		"http://localhost:30301",
		"http://127.0.0.1:30301",
	}
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	return false
}
