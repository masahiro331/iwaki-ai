package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/masahiro331/discord-ai/internal/bot"
	"github.com/masahiro331/discord-ai/internal/discord"
	"github.com/masahiro331/discord-ai/internal/summarizer"
)

const summarizeCmdName = "summarize"

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		return errors.New("DISCORD_BOT_TOKEN env var is required")
	}
	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		return errors.New("DISCORD_GUILD_ID env var is required (guild-scoped commands)")
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return errors.New("GEMINI_API_KEY env var is required")
	}

	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return fmt.Errorf("init discordgo: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	api := discord.NewDiscordgoAPIWithSession(session)
	fetcher := discord.NewFetcher(api)
	llm := summarizer.NewGeminiClient(apiKey)
	sum := summarizer.New(llm)

	handler := bot.NewHandler(fetcher, sum, bot.HandlerConfig{MaxInputChars: 100000})

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		data := i.ApplicationCommandData()
		if data.Name != summarizeCmdName {
			return
		}
		handleSummarize(s, i, handler)
	})

	if err := session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}
	defer session.Close()

	registered, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, &discordgo.ApplicationCommand{
		Name:        summarizeCmdName,
		Description: "Summarize recent messages in this channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "since",
				Description: "How far back to look (e.g. 24h, 3h). Default 24h.",
				Required:    false,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("register command: %w", err)
	}
	log.Printf("registered /%s on guild %s", registered.Name, guildID)

	log.Println("bot is running. Ctrl+C to stop.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
	return nil
}

func handleSummarize(s *discordgo.Session, i *discordgo.InteractionCreate, h *bot.Handler) {
	sinceArg := ""
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "since" {
			sinceArg = opt.StringValue()
		}
	}
	since, err := bot.ParseSinceArg(sinceArg)
	if err != nil {
		replyEphemeral(s, i, fmt.Sprintf("invalid since: %v", err))
		return
	}

	// Defer response: real work takes longer than 3s.
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Printf("defer interaction: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	summary, err := h.Run(ctx, bot.Request{ChannelID: i.ChannelID, Since: since})
	if err != nil {
		msg := fmt.Sprintf("failed: %v", err)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	chunks := discord.SplitForDiscord(summary, discord.DiscordMessageLimit)
	if len(chunks) == 0 {
		empty := "(empty summary)"
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &empty})
		return
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &chunks[0]}); err != nil {
		log.Printf("edit response: %v", err)
		return
	}
	for _, chunk := range chunks[1:] {
		if _, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{Content: chunk}); err != nil {
			log.Printf("followup: %v", err)
			return
		}
	}
}

func replyEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
