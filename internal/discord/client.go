package discord

import (
	"context"
	"fmt"

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

// NewDiscordgoAPIWithSession wraps a pre-built session, useful when the
// caller (e.g. a Bot) needs to register handlers on it before opening.
func NewDiscordgoAPIWithSession(s *discordgo.Session) *DiscordgoAPI {
	return &DiscordgoAPI{session: s}
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
		ts := m.Timestamp
		bot := false
		if m.Author != nil {
			bot = m.Author.Bot
		}
		out = append(out, RawMessage{
			ID:        m.ID,
			Author:    displayNameFromMessage(m),
			Content:   m.Content,
			Timestamp: ts,
			Bot:       bot,
		})
	}
	return out, nil
}

// displayNameFromMessage picks the most human-friendly name available
// for the author. Server nickname wins, then the account's global
// display name, then the legacy username, then a literal "unknown"
// so we never end up with the bare snowflake ID.
func displayNameFromMessage(m *discordgo.Message) string {
	if m.Member != nil && m.Member.Nick != "" {
		return m.Member.Nick
	}
	if m.Author != nil {
		if m.Author.GlobalName != "" {
			return m.Author.GlobalName
		}
		if m.Author.Username != "" {
			return m.Author.Username
		}
	}
	return "unknown"
}
