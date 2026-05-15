package main

import (
	"testing"
	"time"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := parseFlags([]string{"--channel", "123"})
	if err != nil {
		t.Fatalf("parseFlags error = %v", err)
	}
	if cfg.channelID != "123" {
		t.Errorf("channelID = %q, want 123", cfg.channelID)
	}
	if cfg.since != 24*time.Hour {
		t.Errorf("since default = %v, want 24h", cfg.since)
	}
	if cfg.llm != llmClaudeCode {
		t.Errorf("llm default = %q, want %q", cfg.llm, llmClaudeCode)
	}
}

func TestParseFlags_LLMAPI(t *testing.T) {
	cfg, err := parseFlags([]string{"--channel", "x", "--llm", "api"})
	if err != nil {
		t.Fatalf("parseFlags error = %v", err)
	}
	if cfg.llm != llmAPI {
		t.Errorf("llm = %q, want %q", cfg.llm, llmAPI)
	}
}

func TestParseFlags_LLMGemini(t *testing.T) {
	cfg, err := parseFlags([]string{"--channel", "x", "--llm", "gemini"})
	if err != nil {
		t.Fatalf("parseFlags error = %v", err)
	}
	if cfg.llm != llmGemini {
		t.Errorf("llm = %q, want %q", cfg.llm, llmGemini)
	}
}

func TestParseFlags_InvalidLLM(t *testing.T) {
	_, err := parseFlags([]string{"--channel", "x", "--llm", "gpt"})
	if err == nil {
		t.Fatal("expected error on unknown --llm value")
	}
}

func TestParseFlags_CustomSince(t *testing.T) {
	cfg, err := parseFlags([]string{"--channel", "abc", "--since", "3h"})
	if err != nil {
		t.Fatalf("parseFlags error = %v", err)
	}
	if cfg.since != 3*time.Hour {
		t.Errorf("since = %v, want 3h", cfg.since)
	}
}

func TestParseFlags_MissingChannel(t *testing.T) {
	_, err := parseFlags([]string{})
	if err == nil {
		t.Fatal("expected error when --channel missing")
	}
}

func TestParseFlags_NegativeSince(t *testing.T) {
	_, err := parseFlags([]string{"--channel", "x", "--since", "-1h"})
	if err == nil {
		t.Fatal("expected error on negative since")
	}
}
