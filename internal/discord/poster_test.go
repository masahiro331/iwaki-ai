package discord

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakePoster struct {
	calls []fakePostCall
	err   error
}

type fakePostCall struct {
	channelID string
	content   string
}

func (f *fakePoster) PostMessage(_ context.Context, channelID, content string) error {
	f.calls = append(f.calls, fakePostCall{channelID, content})
	return f.err
}

func TestPostSummary_ShortFitsInOneCall(t *testing.T) {
	p := &fakePoster{}
	err := PostSummary(context.Background(), p, "chan-1", "short summary", DiscordMessageLimit)
	if err != nil {
		t.Fatalf("PostSummary error = %v", err)
	}
	if len(p.calls) != 1 {
		t.Fatalf("want 1 call, got %d", len(p.calls))
	}
	if p.calls[0].channelID != "chan-1" || p.calls[0].content != "short summary" {
		t.Errorf("unexpected call: %+v", p.calls[0])
	}
}

func TestPostSummary_LongTextSplits(t *testing.T) {
	p := &fakePoster{}
	// 5 lines of 50 chars = 250+4 newlines; limit=100 forces splitting.
	line := strings.Repeat("a", 50)
	long := strings.Join([]string{line, line, line, line, line}, "\n")

	err := PostSummary(context.Background(), p, "chan-1", long, 100)
	if err != nil {
		t.Fatalf("PostSummary error = %v", err)
	}
	if len(p.calls) < 2 {
		t.Fatalf("expected multiple calls, got %d", len(p.calls))
	}
	for i, c := range p.calls {
		if len(c.content) > 100 {
			t.Errorf("call %d exceeds limit: %d chars", i, len(c.content))
		}
	}
}

func TestPostSummary_PropagatesError(t *testing.T) {
	postErr := errors.New("forbidden")
	p := &fakePoster{err: postErr}

	err := PostSummary(context.Background(), p, "chan-1", "x", DiscordMessageLimit)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, postErr) {
		t.Errorf("error should wrap poster error, got %v", err)
	}
}

func TestPostSummary_EmptyDoesNothing(t *testing.T) {
	p := &fakePoster{}
	if err := PostSummary(context.Background(), p, "chan-1", "", DiscordMessageLimit); err != nil {
		t.Fatalf("PostSummary error = %v", err)
	}
	if len(p.calls) != 0 {
		t.Errorf("empty content should produce no calls, got %d", len(p.calls))
	}
}
