package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"renwen/server/internal/character"
	"renwen/server/internal/config"
	"renwen/server/internal/eval"
	"renwen/server/internal/llm"
	"renwen/server/internal/rag"
)

func main() {
	provider := flag.String("provider", "deepseek", "LLM provider")
	outputDir := flag.String("output-dir", "eval/results", "report output directory")
	charactersFlag := flag.String("characters", "luxun,sushi,lihongzhang,libai,zhugeliang,wangyangming,zhuangzi,wuzetian", "comma-separated character IDs")
	flag.Parse()
	chars, err := character.LoadFromFile(filepath.Join("config", "characters.json"))
	if err != nil {
		panic(err)
	}
	cfg := config.Build(config.LoadAPIKeys(), config.DefaultAppSettings())
	router := llm.NewRouter(*provider, llm.NewDeepSeekProvider(cfg), llm.NewZhihuProvider(cfg))
	retriever, err := rag.NewRetriever()
	if err != nil {
		panic(err)
	}
	report, err := eval.Run(context.Background(), chars, router, retriever, *provider, strings.Split(*charactersFlag, ","))
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		panic(err)
	}
	path := filepath.Join(*outputDir, report.GeneratedAt.Format("20060102T150405Z")+".json")
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		panic(err)
	}
	fmt.Printf("report=%s results=%d severe=%d citations=%d\n", path, len(report.Results), report.SevereCount, report.CitationCount)
}
