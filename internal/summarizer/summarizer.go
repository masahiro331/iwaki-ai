// Package summarizer turns a slice of messages into a natural-language summary
// by delegating to a pluggable LLM client.
package summarizer

import (
	"context"
	"errors"
	"fmt"

	"github.com/masahiro331/discord-ai/internal/message"
)

// ErrNoMessages is returned when there is nothing to summarize.
var ErrNoMessages = errors.New("no messages to summarize")

// LLMClient is the minimal interface a language model implementation must satisfy.
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// Summarizer orchestrates prompt construction and LLM invocation.
type Summarizer struct {
	llm LLMClient
}

// New constructs a Summarizer with the given LLM client.
func New(llm LLMClient) *Summarizer {
	return &Summarizer{llm: llm}
}

// Summarize formats the messages, asks the LLM for a summary, and returns it.
func (s *Summarizer) Summarize(ctx context.Context, msgs []message.Message) (string, error) {
	if len(msgs) == 0 {
		return "", ErrNoMessages
	}
	prompt := buildPrompt(msgs)
	reply, err := s.llm.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("llm complete: %w", err)
	}
	return reply, nil
}

func buildPrompt(msgs []message.Message) string {
	return "以下のDiscordの会話ログを日本語で簡潔に要約してください。" +
		"主な話題、決まったこと、未解決の論点を箇条書きで整理してください。\n\n" +
		"--- 会話ログ ---\n" +
		message.FormatAll(msgs) +
		"\n--- ここまで ---"
}
