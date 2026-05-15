// Package bot wires the existing fetcher and summarizer behind a small
// API that a Discord bot frontend can call. The frontend handles
// discordgo specifics; the logic here stays library-agnostic and tested.
package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/masahiro331/discord-ai/internal/message"
)

// DefaultSinceArg is applied when the user omits the option.
const DefaultSinceArg = 24 * time.Hour

// DefaultMaxSinceArg caps how far back a single command can ask for.
const DefaultMaxSinceArg = 168 * time.Hour

// Fetcher reads messages from Discord. The CLI's discord.Fetcher satisfies it.
type Fetcher interface {
	Fetch(ctx context.Context, channelID string, since time.Time) ([]message.Message, error)
}

// Summarizer turns messages into a summary. The summarizer.Summarizer satisfies it.
type Summarizer interface {
	Summarize(ctx context.Context, msgs []message.Message) (string, error)
}

// Request is the parsed form of a /summarize invocation.
type Request struct {
	ChannelID string
	Since     time.Duration
}

// Result reports the outcome of a /summarize run.
// InputChars and MessageCount are useful for logging and quota tracking.
type Result struct {
	Summary      string
	MessageCount int
	InputChars   int
}

// HandlerConfig holds runtime knobs.
type HandlerConfig struct {
	MaxInputChars int
}

// Handler executes a /summarize request end-to-end.
type Handler struct {
	fetcher    Fetcher
	summarizer Summarizer
	cfg        HandlerConfig
}

// NewHandler constructs a Handler.
func NewHandler(f Fetcher, s Summarizer, cfg HandlerConfig) *Handler {
	if cfg.MaxInputChars <= 0 {
		cfg.MaxInputChars = 100000
	}
	return &Handler{fetcher: f, summarizer: s, cfg: cfg}
}

// ParseSinceArg parses the user-supplied since value (e.g. "3h").
// Empty input falls back to the default. Returns error if invalid,
// non-positive, or exceeding DefaultMaxSinceArg.
func ParseSinceArg(s string) (time.Duration, error) {
	if s == "" {
		return DefaultSinceArg, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("since must be positive (got %s)", d)
	}
	if d > DefaultMaxSinceArg {
		return 0, fmt.Errorf("since %s exceeds maximum %s", d, DefaultMaxSinceArg)
	}
	return d, nil
}

// Run executes the request: fetch -> verify size -> summarize.
func (h *Handler) Run(ctx context.Context, r Request) (Result, error) {
	since := time.Now().Add(-r.Since)
	msgs, err := h.fetcher.Fetch(ctx, r.ChannelID, since)
	if err != nil {
		return Result{}, fmt.Errorf("fetch: %w", err)
	}
	if len(msgs) == 0 {
		return Result{Summary: "(no messages in the given window)"}, nil
	}

	formattedLen := len(message.FormatAll(msgs))
	if formattedLen > h.cfg.MaxInputChars {
		return Result{}, fmt.Errorf("formatted input %d chars exceeds limit %d", formattedLen, h.cfg.MaxInputChars)
	}

	summary, err := h.summarizer.Summarize(ctx, msgs)
	if err != nil {
		return Result{}, fmt.Errorf("summarize: %w", err)
	}
	return Result{
		Summary:      summary,
		MessageCount: len(msgs),
		InputChars:   formattedLen,
	}, nil
}
