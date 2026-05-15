package main

import (
	"flag"
	"fmt"
	"time"
)

type cliConfig struct {
	channelID string
	since     time.Duration
}

func parseFlags(args []string) (cliConfig, error) {
	fs := flag.NewFlagSet("discord-ai-cli", flag.ContinueOnError)
	var cfg cliConfig
	fs.StringVar(&cfg.channelID, "channel", "", "Discord channel ID to summarize (required)")
	fs.DurationVar(&cfg.since, "since", 24*time.Hour, "How far back to look (e.g. 24h, 3h, 7d-style: 168h)")

	if err := fs.Parse(args); err != nil {
		return cliConfig{}, err
	}
	if cfg.channelID == "" {
		return cliConfig{}, fmt.Errorf("--channel is required")
	}
	if cfg.since <= 0 {
		return cliConfig{}, fmt.Errorf("--since must be positive")
	}
	return cfg, nil
}
