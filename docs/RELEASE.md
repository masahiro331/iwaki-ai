# Releases

Tag-driven via GoReleaser. Every `v*` tag triggers
`.github/workflows/release.yml`, which cross-compiles the bot and
CLI and publishes a GitHub Release with the binaries.

## Cutting a release

```bash
git checkout main
git pull
git tag v0.x.y          # follow semver
git push origin v0.x.y
```

The workflow runs in ~2 minutes and uploads:

- `iwaki-ai-bot_<version>_linux_amd64.tar.gz`
- `iwaki-ai-bot_<version>_linux_arm64.tar.gz`
- `iwaki-ai-cli_<version>_{linux,darwin}_{amd64,arm64}.tar.gz`
- `checksums.txt`

`scripts/install.sh` on the VM auto-detects the host architecture
and downloads the matching `iwaki-ai-bot_*` archive from the
**latest** release tag, so as soon as the workflow finishes a new
tag is what cloud-init and `update.sh` will pick up.

## Updating an already-deployed VM

```bash
ssh -i ~/.ssh/iwaki-ai ubuntu@<public_ip>
curl -fsSL https://raw.githubusercontent.com/masahiro331/iwaki-ai/main/scripts/update.sh \
  | sudo bash
```

`update.sh` is idempotent: it reads the binary's `-version` output,
skips the download when it's already on the latest tag, and only
restarts `iwaki-ai-bot.service` when the binary actually changed.

## Versioning notes

- Pre-1.0: `v0.x.y`, `x` for noticeable changes, `y` for patches.
- Once stable: standard semver. Breaking config changes (env vars,
  flag names) belong in major bumps.
- The bot's `-version` output is wired through `-ldflags` in
  `.goreleaser.yaml`; un-stamped local builds report `dev`.

## Rolling back

If a release ships a regression:

1. Re-tag the previous good commit as a new patch
   (`v0.1.4` → `v0.1.5` on the v0.1.3 SHA), so the latest-release
   query picks it up.
2. Run `update.sh` on the VM; the binary swap restarts the service.

Do not delete a published tag/release — that confuses
`update.sh` (it'll keep redownloading) and any external links.
