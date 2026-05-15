# Development

How to read the codebase and run the bot or CLI locally.

## Repository layout

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
terraform/    Oracle Cloud Always Free provisioning
              (also hosts deploy/systemd/ unit file fetched by install.sh)
docs/         You are here
```

The handler and LLM clients are kept behind small interfaces so the
TDD loop works without a Discord or Gemini account; see
`internal/bot/handler_test.go` for the wiring.

## Local dev loop

```bash
make test       # go test ./...
make fmt        # gofmt -s -w .
make vet        # go vet ./...
make build      # bin/iwaki-ai-cli
make build-bot  # bin/iwaki-ai-bot
```

Linting matches the GitHub Actions workflow:

```bash
golangci-lint run ./...
```

## Running the CLI

```bash
export DISCORD_BOT_TOKEN=...
export GEMINI_API_KEY=...                    # or use --llm claude-code
go run ./cmd/cli --channel <CHANNEL_ID> --since 24h --llm gemini
```

Useful flags: `--since`, `--max-since`, `--max-input-chars`,
`--model`, `--no-post`, `--post-to`. Run with `--help` for the full
list.

## Running the bot

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_GUILD_ID=...                  # guild-scoped command
export GEMINI_API_KEY=...
make run-bot
```

Then in the configured guild: `/summarize` (defaults to the last 24
hours) or `/summarize since:3h`.

## LLM backend selection

Three implementations live under `internal/summarizer/`:

- `gemini` (default for the bot): Google AI Studio, free tier
- `api`: Anthropic Messages API; requires `ANTHROPIC_API_KEY`
- `claude-code`: shells out to the local `claude -p` binary; works
  while a Claude Code Max session is signed in

The CLI picks the backend with `--llm`. The bot wires the primary
through `RetryClient` with a `gemini-2.0-flash` fallback for 5xx /
429 responses; see `cmd/bot/main.go`.

## Adding a new LLM backend

1. Add a struct implementing `LLMClient` (`Complete(ctx, prompt)
   (string, error)`) under `internal/summarizer/`.
2. Wrap retryable transport errors in `&RetryableError{Cause: ...}`
   so `RetryClient` will back off and try again.
3. In the CLI, extend `cmd/cli/llm.go` to recognize the new
   `--llm` value.
