# CLAUDE.md

Working notes for AI assistants (and humans) collaborating on this
repo. The repo is developed **issue-driven**: every non-trivial
change starts from a GitHub Issue, and the README does not maintain
a separate roadmap that would inevitably drift.

## How work flows here

1. **Ideas live in Issues**, not in README TODOs or planning docs.
   - Bug? Open an issue.
   - New feature? Open an issue.
   - Refactor or chore worth tracking? Open an issue.
   - If you're tempted to write "TODO: …" in a markdown file, write
     an issue instead and link it from the relevant doc/code.

2. **Branch + PR per issue.**
   - One issue → one branch named `feat/...`, `fix/...`, `chore/...`,
     `docs/...` matching the change type.
   - PR body should reference the issue (`Closes #123`).

3. **Issue scope.**
   - One issue = one shippable outcome. Break large work into
     sub-issues rather than one mega-issue with checkboxes.
   - When an issue's specification isn't nailed down, capture the
     open questions inside the issue itself (the templates already
     have a "仕様未確定の点" section).

## Coding conventions

- Comments and commit messages: **English**.
- User-facing strings in the bot (replies, error messages) and the
  end-user documentation under `docs/USER_GUIDE.ja.md`: **Japanese**.
- Identifiers (functions, types, vars): English, descriptive.
- Tests are written first (TDD); see `internal/bot/handler_test.go`
  for the style.
- Run `make fmt && make vet && make test && golangci-lint run ./...`
  before pushing.

## What not to do

- Don't add a "Planned features" or "Roadmap" section to any
  README; link to the issue list instead.
- Don't paste Claude attribution lines (`Co-Authored-By: Claude …`)
  into commit messages or PR bodies; CI rejects them.
- Don't widen scope inside a PR. If a tangentially related cleanup
  comes up, file a new issue and reference it.

## Where things live

- `cmd/bot/` — long-running Discord bot
- `cmd/cli/` — one-shot CLI (mainly for local dev / testing)
- `internal/bot/` — slash-command handler, discordgo-free, fully tested
- `internal/discord/` — discordgo wrapper (fetch, post, split, …)
- `internal/summarizer/` — pluggable LLM backends behind one interface
- `terraform/` — Oracle Cloud Always Free provisioning
- `scripts/install.sh`, `scripts/update.sh` — VM-side bootstrap/update
- `docs/` — all long-form docs (deployment, release, user guide, …)
