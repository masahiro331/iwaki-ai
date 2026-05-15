package main

import (
	"fmt"
	"os"

	"github.com/masahiro331/discord-ai/internal/summarizer"
)

// buildLLM constructs an LLMClient based on the CLI configuration.
// "claude-code" uses the local `claude -p` binary (no API key required).
// "api" uses the Anthropic Messages API and requires ANTHROPIC_API_KEY.
// "gemini" uses the Google AI Studio API and requires GEMINI_API_KEY.
func buildLLM(cfg cliConfig) (summarizer.LLMClient, error) {
	switch cfg.llm {
	case llmClaudeCode:
		return summarizer.NewClaudeCodeClient(), nil
	case llmAPI:
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY env var is required when --llm=%s", llmAPI)
		}
		var opts []summarizer.ClaudeOption
		if cfg.model != "" {
			opts = append(opts, summarizer.WithModel(cfg.model))
		}
		return summarizer.NewClaudeClient(apiKey, opts...), nil
	case llmGemini:
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY env var is required when --llm=%s", llmGemini)
		}
		var opts []summarizer.GeminiOption
		if cfg.model != "" {
			opts = append(opts, summarizer.WithGeminiModel(cfg.model))
		}
		return summarizer.NewGeminiClient(apiKey, opts...), nil
	default:
		return nil, fmt.Errorf("unknown llm kind: %q", cfg.llm)
	}
}
