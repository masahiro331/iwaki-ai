package discord

import (
	"context"
	"fmt"
)

// Poster sends a single message to a Discord channel.
type Poster interface {
	PostMessage(ctx context.Context, channelID, content string) error
}

// PostSummary sends `content` to `channelID`, splitting into multiple
// messages if it exceeds `limit` characters. No-op for empty content.
func PostSummary(ctx context.Context, p Poster, channelID, content string, limit int) error {
	chunks := SplitForDiscord(content, limit)
	for i, chunk := range chunks {
		if err := p.PostMessage(ctx, channelID, chunk); err != nil {
			return fmt.Errorf("post chunk %d/%d: %w", i+1, len(chunks), err)
		}
	}
	return nil
}
