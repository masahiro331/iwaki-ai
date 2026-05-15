package main

import (
	"fmt"
	"os"

	"github.com/masahiro331/discord-ai/internal/summarizer"
)

// buildLLM constructs an LLMClient based on the --llm flag value.
// "claude-code" uses the local `claude -p` binary (no API key required).
// "api" uses the Anthropic Messages API and requires ANTHROPIC_API_KEY.
func buildLLM(kind string) (summarizer.LLMClient, error) {
	switch kind {
	case llmClaudeCode:
		return summarizer.NewClaudeCodeClient(), nil
	case llmAPI:
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY env var is required when --llm=%s", llmAPI)
		}
		return summarizer.NewClaudeClient(apiKey), nil
	default:
		return nil, fmt.Errorf("unknown llm kind: %q", kind)
	}
}
