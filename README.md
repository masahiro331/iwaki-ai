# Iwaki AI

A Discord bot that summarizes recent conversation in a channel.
Members type `/summarize` (optionally `since:3h`, `since:24h`, …)
and the bot replies with a Markdown-style digest produced by an
LLM backend (Google Gemini by default).

> 利用者向けの日本語ガイドは [docs/USER_GUIDE.ja.md](docs/USER_GUIDE.ja.md)。

## Docs

| Audience | Read |
|---|---|
| Discord users running `/summarize` | [docs/USER_GUIDE.ja.md](docs/USER_GUIDE.ja.md) |
| Developers hacking on the code | [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) |
| Operators deploying to Oracle Cloud | [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) |
| Maintainers cutting releases | [docs/RELEASE.md](docs/RELEASE.md) |

## Quick taste

```bash
# Local CLI run (one-shot summary printed to stdout)
export DISCORD_BOT_TOKEN=...
export GEMINI_API_KEY=...
go run ./cmd/cli --channel <CHANNEL_ID> --since 24h --llm gemini --no-post
```

```bash
# Local bot (registers /summarize against $DISCORD_GUILD_ID)
export DISCORD_BOT_TOKEN=...
export DISCORD_GUILD_ID=...
export GEMINI_API_KEY=...
make run-bot
```

For production deploys, see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).
