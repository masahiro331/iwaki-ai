package bot

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/masahiro331/discord-ai/internal/message"
)

type fakeFetcher struct {
	gotChannel string
	gotSince   time.Time
	msgs       []message.Message
	err        error
}

func (f *fakeFetcher) Fetch(_ context.Context, channelID string, since time.Time) ([]message.Message, error) {
	f.gotChannel = channelID
	f.gotSince = since
	return f.msgs, f.err
}

type fakeSummarizer struct {
	gotMsgs []message.Message
	reply   string
	err     error
}

func (s *fakeSummarizer) Summarize(_ context.Context, msgs []message.Message) (string, error) {
	s.gotMsgs = msgs
	return s.reply, s.err
}

func TestParseSinceArg(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
	}{
		{"", 24 * time.Hour}, // default
		{"3h", 3 * time.Hour},
		{"168h", 168 * time.Hour},
	}
	for _, tt := range tests {
		got, err := ParseSinceArg(tt.in)
		if err != nil {
			t.Errorf("ParseSinceArg(%q) error = %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseSinceArg(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseSinceArg_Invalid(t *testing.T) {
	cases := []string{"abc", "-1h", "0h", "300h"} // 300h exceeds default cap 168h
	for _, in := range cases {
		_, err := ParseSinceArg(in)
		if err == nil {
			t.Errorf("ParseSinceArg(%q) expected error, got nil", in)
		}
	}
}

func TestSummarizeRequest_HappyPath(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	msgs := []message.Message{
		{Author: "a", Content: "hi", Timestamp: time.Date(2026, 5, 16, 10, 0, 0, 0, jst)},
	}
	f := &fakeFetcher{msgs: msgs}
	s := &fakeSummarizer{reply: "summary"}

	h := NewHandler(f, s, HandlerConfig{MaxInputChars: 1000})

	got, err := h.Run(context.Background(), Request{
		ChannelID: "chan-1",
		Since:     3 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got.Summary != "summary" {
		t.Errorf("Run() summary = %q, want %q", got.Summary, "summary")
	}
	if got.MessageCount != 1 {
		t.Errorf("Run() MessageCount = %d, want 1", got.MessageCount)
	}
	if got.InputChars <= 0 {
		t.Errorf("Run() InputChars = %d, want > 0", got.InputChars)
	}
	if f.gotChannel != "chan-1" {
		t.Errorf("fetcher channel = %q, want chan-1", f.gotChannel)
	}
	// Since should be roughly 3 hours before now.
	delta := time.Since(f.gotSince)
	if delta < 2*time.Hour+50*time.Minute || delta > 3*time.Hour+10*time.Minute {
		t.Errorf("fetcher since delta = %v, want ~3h", delta)
	}
}

func TestSummarizeRequest_NoMessages(t *testing.T) {
	f := &fakeFetcher{msgs: nil}
	s := &fakeSummarizer{reply: "should not be called"}

	h := NewHandler(f, s, HandlerConfig{MaxInputChars: 1000})

	got, err := h.Run(context.Background(), Request{
		ChannelID: "c", Since: time.Hour,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(got.Summary, "no messages") && !strings.Contains(got.Summary, "メッセージ") {
		t.Errorf("Run() should explain no messages, got %q", got.Summary)
	}
	if got.MessageCount != 0 {
		t.Errorf("Run() MessageCount = %d, want 0", got.MessageCount)
	}
	if s.gotMsgs != nil {
		t.Errorf("summarizer should not be called when no messages")
	}
}

func TestSummarizeRequest_FetcherError(t *testing.T) {
	fErr := errors.New("403")
	f := &fakeFetcher{err: fErr}
	s := &fakeSummarizer{}

	h := NewHandler(f, s, HandlerConfig{MaxInputChars: 1000})

	_, err := h.Run(context.Background(), Request{ChannelID: "c", Since: time.Hour})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, fErr) {
		t.Errorf("error should wrap fetcher error, got %v", err)
	}
}

func TestSummarizeRequest_InputTooLarge(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	long := strings.Repeat("x", 5000)
	msgs := []message.Message{
		{Author: "a", Content: long, Timestamp: time.Date(2026, 5, 16, 10, 0, 0, 0, jst)},
	}
	f := &fakeFetcher{msgs: msgs}
	s := &fakeSummarizer{reply: "should not be called"}

	h := NewHandler(f, s, HandlerConfig{MaxInputChars: 1000})

	_, err := h.Run(context.Background(), Request{ChannelID: "c", Since: time.Hour})
	if err == nil {
		t.Fatal("expected error on oversize input, got nil")
	}
	if s.gotMsgs != nil {
		t.Errorf("summarizer should not be called when input exceeds limit")
	}
}
