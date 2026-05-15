// Package discord fetches messages from Discord channels.
package discord

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/masahiro331/iwaki-ai/internal/message"
)

// defaultPageLimit is the maximum number of messages requested per API call.
// Discord caps this at 100.
const defaultPageLimit = 100

// RawMessage is the minimal subset of fields we need from a Discord message.
type RawMessage struct {
	ID        string
	Author    string
	Content   string
	Timestamp time.Time
	Bot       bool
}

// MessagesAPI is the low-level interface a Discord client must satisfy.
// It returns up to `limit` messages older than `beforeID`, newest-first.
// When beforeID is empty, it returns the latest messages.
type MessagesAPI interface {
	ChannelMessages(ctx context.Context, channelID, beforeID string, limit int) ([]RawMessage, error)
}

// Fetcher collects messages within a time window across paginated API calls.
type Fetcher struct {
	api       MessagesAPI
	pageLimit int
}

// FetcherOption customizes a Fetcher.
type FetcherOption func(*Fetcher)

// WithPageLimit overrides the per-request message limit (mainly for tests).
func WithPageLimit(n int) FetcherOption {
	return func(f *Fetcher) { f.pageLimit = n }
}

// NewFetcher builds a Fetcher backed by the given MessagesAPI.
func NewFetcher(api MessagesAPI, opts ...FetcherOption) *Fetcher {
	f := &Fetcher{api: api, pageLimit: defaultPageLimit}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Fetch returns messages in `channelID` whose timestamp is at or after `since`,
// excluding bot messages, in chronological (oldest-first) order.
func (f *Fetcher) Fetch(ctx context.Context, channelID string, since time.Time) ([]message.Message, error) {
	var collected []message.Message
	beforeID := ""

	for {
		page, err := f.api.ChannelMessages(ctx, channelID, beforeID, f.pageLimit)
		if err != nil {
			return nil, fmt.Errorf("fetch channel %s: %w", channelID, err)
		}
		if len(page) == 0 {
			break
		}

		reachedSince := false
		for _, raw := range page {
			if raw.Timestamp.Before(since) {
				reachedSince = true
				continue
			}
			if raw.Bot {
				continue
			}
			collected = append(collected, message.Message{
				Author:    raw.Author,
				Content:   raw.Content,
				Timestamp: raw.Timestamp,
			})
		}

		if reachedSince || len(page) < f.pageLimit {
			break
		}
		// Advance pagination cursor: oldest ID in this page (page is newest-first).
		beforeID = page[len(page)-1].ID
	}

	sort.Slice(collected, func(i, j int) bool {
		return collected[i].Timestamp.Before(collected[j].Timestamp)
	})
	return collected, nil
}
