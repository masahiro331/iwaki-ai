# Iwaki AI

Discord conversation summarizer. A `/summarize` slash command pulls
recent messages from the channel it's invoked in, hands them to
Google's Gemini API, and posts the result back to the channel.

Built mainly so the operator can catch up on busy channels without
scrolling.

> Looking for the user-facing Bot guide in 日本語? See
> [README.ja.md](README.ja.md).

## What's here

```
cmd/
  bot/        Long-running Discord bot exposing /summarize
  cli/        One-shot CLI for ad-hoc summarization
internal/
  bot/        /summarize handler (discordgo-free, unit tested)
  discord/    discordgo wrapper: fetcher, poster, paging, 2000-char split
  message/    Common Message type + prompt formatting
  summarizer/ LLMClient interface plus Claude/Gemini/Claude-Code impls,
              retry-with-fallback wrapper
scripts/      install.sh / update.sh helpers run on the VM
terraform/    Oracle Cloud Always Free provisioning (see its README);
              also hosts deploy/systemd/ unit file fetched by install.sh
```

## CLI

```bash
export DISCORD_BOT_TOKEN=...
export GEMINI_API_KEY=...                    # or use --llm claude-code
go run ./cmd/cli --channel <CHANNEL_ID> --since 24h --llm gemini
```

Useful flags: `--since`, `--max-since`, `--max-input-chars`,
`--model`, `--no-post`, `--post-to`. Run with `--help` for the full
list.

## Bot

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_GUILD_ID=...                  # guild-scoped command
export GEMINI_API_KEY=...
make run-bot
```

Then in the configured guild: `/summarize` (defaults to the last 24
hours) or `/summarize since:3h`.

## Deploying to Oracle Cloud

See [terraform/README.md](terraform/README.md) for the one-time
setup (OCI CLI, SOPS+age, Discord/Gemini secrets) and the
`terraform apply` flow that brings up an Always Free A1.Flex VM
running the bot under systemd.

## Releases

Tag-driven via GoReleaser:

```bash
git tag v0.x.y
git push origin v0.x.y
```

GitHub Release publishes `iwaki-ai-bot_*.tar.gz` (linux amd64/arm64)
and `iwaki-ai-cli_*.tar.gz` (linux + darwin, amd64/arm64) plus a
`checksums.txt`. `scripts/install.sh` on the VM pulls the matching
bot archive for the host architecture.

## Local dev loop

```bash
make test       # go test ./...
make fmt        # gofmt -s -w .
make vet        # go vet ./...
make build      # bin/iwaki-ai-cli
make build-bot  # bin/iwaki-ai-bot
```
