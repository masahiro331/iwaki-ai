package main

import (
	"flag"
	"fmt"
	"time"
)

const (
	llmClaudeCode = "claude-code"
	llmAPI        = "api"
)

type cliConfig struct {
	channelID string
	since     time.Duration
	llm       string
}

func parseFlags(args []string) (cliConfig, error) {
	fs := flag.NewFlagSet("discord-ai-cli", flag.ContinueOnError)
	var cfg cliConfig
	fs.StringVar(&cfg.channelID, "channel", "", "Discord channel ID to summarize (required)")
	fs.DurationVar(&cfg.since, "since", 24*time.Hour, "How far back to look (e.g. 24h, 3h, 7d-style: 168h)")
	fs.StringVar(&cfg.llm, "llm", llmClaudeCode, "LLM backend: claude-code (uses local `claude -p`) or api (Anthropic API)")

	if err := fs.Parse(args); err != nil {
		return cliConfig{}, err
	}
	if cfg.channelID == "" {
		return cliConfig{}, fmt.Errorf("--channel is required")
	}
	if cfg.since <= 0 {
		return cliConfig{}, fmt.Errorf("--since must be positive")
	}
	if cfg.llm != llmClaudeCode && cfg.llm != llmAPI {
		return cliConfig{}, fmt.Errorf("--llm must be %q or %q (got %q)", llmClaudeCode, llmAPI, cfg.llm)
	}
	return cfg, nil
}
