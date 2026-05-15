# Iwaki AI

A Discord bot for a **Nobunaga's Ambition: Awakening (信長の野望 真戦)**
community server. The current command is `/summarize` (recap recent
channel chatter via an LLM); more game-specific features are on the
way - alliance member tracking, schedule reminders, siege/march
planning helpers, lookups for events / units / commanders, an
ingest of the game's official X posts, and eventually image
understanding so screenshots can be parsed too.

> 利用者向けの日本語ガイドは [docs/USER_GUIDE.ja.md](docs/USER_GUIDE.ja.md)。

## Docs

| Audience | Read |
|---|---|
| Discord users running commands | [docs/USER_GUIDE.ja.md](docs/USER_GUIDE.ja.md) |
| Developers hacking on the code | [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) |
| Operators deploying to Oracle Cloud | [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) |
| Maintainers cutting releases | [docs/RELEASE.md](docs/RELEASE.md) |

## Shipped commands

| Command | What it does |
|---|---|
| `/summarize [since:24h]` | Summarize recent channel messages with Gemini |

Planned features are tracked as GitHub Issues. New ideas → open an
issue rather than editing a TODO list here:
<https://github.com/masahiro331/iwaki-ai/issues>.

This repo is developed **issue-driven** — see [CLAUDE.md](CLAUDE.md)
for how that works in practice.

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
