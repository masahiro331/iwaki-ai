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
