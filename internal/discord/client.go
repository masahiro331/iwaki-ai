package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// DiscordgoAPI adapts a *discordgo.Session to the MessagesAPI interface.
type DiscordgoAPI struct {
	session *discordgo.Session
}

// NewDiscordgoAPI builds a MessagesAPI backed by discordgo using the given bot token.
func NewDiscordgoAPI(token string) (*DiscordgoAPI, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("init discordgo: %w", err)
	}
	return &DiscordgoAPI{session: s}, nil
}

// PostMessage sends a plain text message to the given channel.
func (a *DiscordgoAPI) PostMessage(ctx context.Context, channelID, content string) error {
	_, err := a.session.ChannelMessageSend(channelID, content, discordgo.WithContext(ctx))
	return err
}

// ChannelMessages fetches up to `limit` messages older than beforeID, newest-first.
func (a *DiscordgoAPI) ChannelMessages(ctx context.Context, channelID, beforeID string, limit int) ([]RawMessage, error) {
	msgs, err := a.session.ChannelMessages(channelID, limit, beforeID, "", "", discordgo.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]RawMessage, 0, len(msgs))
	for _, m := range msgs {
		ts := time.Time(m.Timestamp)
		author := ""
		bot := false
		if m.Author != nil {
			author = m.Author.Username
			bot = m.Author.Bot
		}
		out = append(out, RawMessage{
			ID:        m.ID,
			Author:    author,
			Content:   m.Content,
			Timestamp: ts,
			Bot:       bot,
		})
	}
	return out, nil
}
