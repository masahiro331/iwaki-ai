package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masahiro331/iwaki-ai/internal/message"
)

// fakeAPI returns canned messages in a single page (or splits if you want).
type fakeAPI struct {
	pages [][]RawMessage // ordered pages newest-first as returned by Discord
	calls []fakeAPICall
	err   error
}

type fakeAPICall struct {
	channelID string
	beforeID  string
	limit     int
}

func (f *fakeAPI) ChannelMessages(_ context.Context, channelID, beforeID string, limit int) ([]RawMessage, error) {
	f.calls = append(f.calls, fakeAPICall{channelID, beforeID, limit})
	if f.err != nil {
		return nil, f.err
	}
	if len(f.pages) == 0 {
		return nil, nil
	}
	page := f.pages[0]
	f.pages = f.pages[1:]
	return page, nil
}

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestFetcher_Fetch_FiltersBySince(t *testing.T) {
	since := mustTime("2026-05-15T00:00:00Z")
	api := &fakeAPI{
		pages: [][]RawMessage{
			// Discord returns newest first
			{
				{ID: "3", Author: "alice", Content: "newer", Timestamp: mustTime("2026-05-15T12:00:00Z"), Bot: false},
				{ID: "2", Author: "bob", Content: "edge", Timestamp: mustTime("2026-05-15T00:00:00Z"), Bot: false},
				{ID: "1", Author: "carol", Content: "older", Timestamp: mustTime("2026-05-14T23:00:00Z"), Bot: false},
			},
		},
	}

	f := NewFetcher(api, WithPageLimit(2))
	got, err := f.Fetch(context.Background(), "chan-1", since)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Older message must be excluded; "edge" (at since) kept (inclusive)
	if len(got) != 2 {
		t.Fatalf("want 2 messages, got %d: %+v", len(got), got)
	}
	// Expect chronological order (oldest first)
	if got[0].Author != "bob" || got[1].Author != "alice" {
		t.Errorf("messages not in chronological order: %+v", got)
	}
	if _, ok := any(got[0]).(message.Message); !ok {
		t.Errorf("expected []message.Message")
	}
}

func TestFetcher_Fetch_Pagination(t *testing.T) {
	since := mustTime("2026-05-15T00:00:00Z")
	api := &fakeAPI{
		pages: [][]RawMessage{
			{
				{ID: "5", Author: "u5", Content: "m5", Timestamp: mustTime("2026-05-15T12:00:00Z")},
				{ID: "4", Author: "u4", Content: "m4", Timestamp: mustTime("2026-05-15T11:00:00Z")},
			},
			{
				{ID: "3", Author: "u3", Content: "m3", Timestamp: mustTime("2026-05-15T10:00:00Z")},
				{ID: "2", Author: "u2", Content: "m2", Timestamp: mustTime("2026-05-15T09:00:00Z")},
			},
			{
				// last page contains a message older than `since` -> stop after this
				{ID: "1", Author: "u1", Content: "m1", Timestamp: mustTime("2026-05-14T23:00:00Z")},
			},
		},
	}

	f := NewFetcher(api, WithPageLimit(2))
	got, err := f.Fetch(context.Background(), "chan-1", since)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("want 4 messages, got %d", len(got))
	}
	// Verify "before" cursor is forwarded between pages
	if len(api.calls) < 2 {
		t.Fatalf("expected paginated calls, got %d", len(api.calls))
	}
	if api.calls[0].beforeID != "" {
		t.Errorf("first call should have empty beforeID, got %q", api.calls[0].beforeID)
	}
	if api.calls[1].beforeID != "4" {
		t.Errorf("second call beforeID = %q, want last ID of prior page (4)", api.calls[1].beforeID)
	}
}

func TestFetcher_Fetch_SkipsBots(t *testing.T) {
	since := mustTime("2026-05-15T00:00:00Z")
	api := &fakeAPI{
		pages: [][]RawMessage{
			{
				{ID: "2", Author: "bot", Content: "hi", Timestamp: mustTime("2026-05-15T10:00:00Z"), Bot: true},
				{ID: "1", Author: "alice", Content: "real", Timestamp: mustTime("2026-05-15T09:00:00Z"), Bot: false},
			},
		},
	}
	f := NewFetcher(api, WithPageLimit(2))
	got, err := f.Fetch(context.Background(), "chan-1", since)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(got) != 1 || got[0].Author != "alice" {
		t.Errorf("bot messages should be filtered; got %+v", got)
	}
}

func TestFetcher_Fetch_APIError(t *testing.T) {
	apiErr := errors.New("network down")
	api := &fakeAPI{err: apiErr}
	f := NewFetcher(api, WithPageLimit(2))

	_, err := f.Fetch(context.Background(), "chan-1", time.Now().Add(-time.Hour))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apiErr) {
		t.Errorf("want wrapped api error, got %v", err)
	}
}
