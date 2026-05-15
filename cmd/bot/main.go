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

	"github.com/masahiro331/iwaki-ai/internal/bot"
	"github.com/masahiro331/iwaki-ai/internal/discord"
	"github.com/masahiro331/iwaki-ai/internal/summarizer"
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

	primary := summarizer.NewGeminiClient(apiKey) // gemini-2.5-flash by default
	fallback := summarizer.NewGeminiClient(apiKey, summarizer.WithGeminiModel("gemini-2.0-flash"))
	llm := summarizer.NewRetryClient(primary, summarizer.RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		Fallback:    fallback,
	})
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
	log.Printf("registered slash command name=%q guild=%q", registered.Name, guildID)

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
	user, userID := userIdentity(i)
	log.Printf("/summarize invoked: user=%s (id=%s) guild=%s channel=%s since=%q",
		user, userID, i.GuildID, i.ChannelID, sinceArg)

	since, err := bot.ParseSinceArg(sinceArg)
	if err != nil {
		log.Printf("invalid since from user=%s: %v", user, err)
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

	start := time.Now()
	result, err := h.Run(ctx, bot.Request{ChannelID: i.ChannelID, Since: since})
	if err != nil {
		log.Printf("/summarize failed for user=%s channel=%s: %v", user, i.ChannelID, err)
		msg := fmt.Sprintf("failed: %v", err)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}
	log.Printf("/summarize succeeded for user=%s channel=%s since=%s elapsed=%s messages=%d input_chars=%d summary_chars=%d",
		user, i.ChannelID, since, time.Since(start).Round(time.Millisecond),
		result.MessageCount, result.InputChars, len(result.Summary))

	chunks := discord.SplitForDiscord(result.Summary, discord.DiscordMessageLimit)
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

// userIdentity returns a printable username and ID, handling both
// guild-member and DM-style invocations.
func userIdentity(i *discordgo.InteractionCreate) (name, id string) {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.Username, i.Member.User.ID
	}
	if i.User != nil {
		return i.User.Username, i.User.ID
	}
	return "unknown", ""
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
