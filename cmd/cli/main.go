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

	llm, err := buildLLM(cfg)
	if err != nil {
		return err
	}

	api, err := discord.NewDiscordgoAPI(botToken)
	if err != nil {
		return fmt.Errorf("init discord client: %w", err)
	}
	fetcher := discord.NewFetcher(api)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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

	s := summarizer.New(llm)

	summary, err := s.Summarize(ctx, msgs)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	fmt.Println(summary)

	if cfg.noPost {
		return nil
	}
	postTarget := cfg.postTo
	if postTarget == "" {
		postTarget = cfg.channelID
	}
	if err := discord.PostSummary(ctx, api, postTarget, summary, discord.DiscordMessageLimit); err != nil {
		return fmt.Errorf("post summary to %s: %w", postTarget, err)
	}
	return nil
}
