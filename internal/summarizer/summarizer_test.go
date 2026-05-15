package summarizer

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/masahiro331/discord-ai/internal/message"
)

type fakeLLM struct {
	gotPrompt string
	reply     string
	err       error
}

func (f *fakeLLM) Complete(_ context.Context, prompt string) (string, error) {
	f.gotPrompt = prompt
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func TestSummarizer_Summarize(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	msgs := []message.Message{
		{Author: "alice", Content: "今日のミーティングは延期", Timestamp: time.Date(2026, 5, 15, 10, 0, 0, 0, jst)},
		{Author: "bob", Content: "了解です", Timestamp: time.Date(2026, 5, 15, 10, 1, 0, 0, jst)},
	}

	llm := &fakeLLM{reply: "ミーティング延期で合意"}
	s := New(llm)

	got, err := s.Summarize(context.Background(), msgs)
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}
	if got != "ミーティング延期で合意" {
		t.Errorf("Summarize() = %q, want LLM reply", got)
	}
	if !strings.Contains(llm.gotPrompt, "alice") || !strings.Contains(llm.gotPrompt, "bob") {
		t.Errorf("prompt should include message authors, got: %q", llm.gotPrompt)
	}
	if !strings.Contains(llm.gotPrompt, "今日のミーティングは延期") {
		t.Errorf("prompt should include message content, got: %q", llm.gotPrompt)
	}
}

func TestSummarizer_Summarize_EmptyMessages(t *testing.T) {
	llm := &fakeLLM{reply: "should not be called"}
	s := New(llm)

	_, err := s.Summarize(context.Background(), nil)
	if err == nil {
		t.Fatal("Summarize(nil) expected error, got nil")
	}
	if llm.gotPrompt != "" {
		t.Errorf("LLM should not be called for empty input, got prompt: %q", llm.gotPrompt)
	}
}

func TestSummarizer_Summarize_LLMError(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	msgs := []message.Message{
		{Author: "alice", Content: "hi", Timestamp: time.Date(2026, 5, 15, 10, 0, 0, 0, jst)},
	}
	llmErr := errors.New("api failure")
	llm := &fakeLLM{err: llmErr}
	s := New(llm)

	_, err := s.Summarize(context.Background(), msgs)
	if err == nil {
		t.Fatal("expected error from LLM, got nil")
	}
	if !errors.Is(err, llmErr) {
		t.Errorf("error should wrap LLM error, got: %v", err)
	}
}
