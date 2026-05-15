package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestDisplayNameFromMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  *discordgo.Message
		want string
	}{
		{
			name: "server nickname wins",
			msg: &discordgo.Message{
				Member: &discordgo.Member{Nick: "たぬきち"},
				Author: &discordgo.User{GlobalName: "Tanuki", Username: "tanukichi331"},
			},
			want: "たぬきち",
		},
		{
			name: "global name when nickname empty",
			msg: &discordgo.Message{
				Member: &discordgo.Member{Nick: ""},
				Author: &discordgo.User{GlobalName: "Tanuki", Username: "tanukichi331"},
			},
			want: "Tanuki",
		},
		{
			name: "username when nickname and global empty",
			msg: &discordgo.Message{
				Author: &discordgo.User{Username: "tanukichi331"},
			},
			want: "tanukichi331",
		},
		{
			name: "no member set falls through to author",
			msg: &discordgo.Message{
				Author: &discordgo.User{GlobalName: "Tanuki"},
			},
			want: "Tanuki",
		},
		{
			name: "nothing usable yields a sentinel rather than empty string",
			msg:  &discordgo.Message{},
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayNameFromMessage(tt.msg); got != tt.want {
				t.Errorf("displayNameFromMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
