package main

import (
	"errors"
	"flag"
	"fmt"
	"time"
)

const (
	llmClaudeCode = "claude-code"
	llmAPI        = "api"
	llmGemini     = "gemini"

	defaultMaxSince      = 168 * time.Hour
	defaultMaxInputChars = 100000
)

type cliConfig struct {
	channelID     string
	since         time.Duration
	maxSince      time.Duration
	llm           string
	model         string
	noPost        bool
	postTo        string
	maxInputChars int
}

func parseFlags(args []string) (cliConfig, error) {
	fs := flag.NewFlagSet("iwaki-ai-cli", flag.ContinueOnError)
	var cfg cliConfig
	fs.StringVar(&cfg.channelID, "channel", "", "Discord channel ID to summarize (required)")
	fs.DurationVar(&cfg.since, "since", 24*time.Hour, "How far back to look (e.g. 24h, 3h, 7d-style: 168h)")
	fs.StringVar(&cfg.llm, "llm", llmClaudeCode, "LLM backend: claude-code (local `claude -p`), api (Anthropic API), or gemini (Google AI Studio)")
	fs.StringVar(&cfg.model, "model", "", "Override the LLM model name (backend-specific, e.g. gemini-2.0-flash). Empty uses the backend default.")
	fs.BoolVar(&cfg.noPost, "no-post", false, "Do not post the summary back to Discord (default is to post)")
	fs.StringVar(&cfg.postTo, "post-to", "", "Discord channel ID to post the summary to (default: same as --channel)")
	fs.DurationVar(&cfg.maxSince, "max-since", defaultMaxSince, "Hard upper bound on --since to prevent accidentally summarizing huge windows")
	fs.IntVar(&cfg.maxInputChars, "max-input-chars", defaultMaxInputChars, "Refuse to summarize when the formatted input exceeds this many characters")

	if err := fs.Parse(args); err != nil {
		return cliConfig{}, err
	}
	if cfg.channelID == "" {
		return cliConfig{}, errors.New("--channel is required")
	}
	if cfg.since <= 0 {
		return cliConfig{}, errors.New("--since must be positive")
	}
	if cfg.since > cfg.maxSince {
		return cliConfig{}, fmt.Errorf("--since %s exceeds --max-since %s; raise --max-since to override", cfg.since, cfg.maxSince)
	}
	if cfg.maxInputChars <= 0 {
		return cliConfig{}, errors.New("--max-input-chars must be positive")
	}
	switch cfg.llm {
	case llmClaudeCode, llmAPI, llmGemini:
		// ok
	default:
		return cliConfig{}, fmt.Errorf("--llm must be one of %q, %q, %q (got %q)",
			llmClaudeCode, llmAPI, llmGemini, cfg.llm)
	}
	return cfg, nil
}
