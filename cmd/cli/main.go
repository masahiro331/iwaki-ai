package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/masahiro331/discord-ai/internal/discord"
	"github.com/masahiro331/discord-ai/internal/summarizer"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := parseFlags(args)
	if err != nil {
		return err
	}

	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		return errors.New("DISCORD_BOT_TOKEN env var is required")
	}
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return errors.New("ANTHROPIC_API_KEY env var is required")
	}

	api, err := discord.NewDiscordgoAPI(botToken)
	if err != nil {
		return fmt.Errorf("init discord client: %w", err)
	}
	fetcher := discord.NewFetcher(api)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	since := time.Now().Add(-cfg.since)
	msgs, err := fetcher.Fetch(ctx, cfg.channelID, since)
	if err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}
	if len(msgs) == 0 {
		fmt.Println("(no messages in the given window)")
		return nil
	}

	llm := summarizer.NewClaudeClient(apiKey)
	s := summarizer.New(llm)

	summary, err := s.Summarize(ctx, msgs)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	fmt.Println(summary)
	return nil
}
