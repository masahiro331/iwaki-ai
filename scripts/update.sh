#!/usr/bin/env bash
# update.sh - swap the running binary for the latest release.
#
# Idempotent: if the host is already on the latest release the script
# exits 0 without restarting the service. Otherwise it downloads the
# new archive, replaces /usr/local/bin/iwaki-ai-bot, and restarts the
# systemd unit so the bot picks up the new build.
#
# Usage:
#   sudo ./update.sh [REPO]
set -euo pipefail

REPO="${1:-masahiro331/iwaki-ai}"
BIN_PATH="/usr/local/bin/iwaki-ai-bot"
SERVICE_NAME="iwaki-ai-bot.service"

if [[ $EUID -ne 0 ]]; then
  echo "update.sh must be run as root (use sudo)" >&2
  exit 1
fi

arch="$(uname -m)"
case "$arch" in
  x86_64)  goarch="amd64" ;;
  aarch64) goarch="arm64" ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

echo "fetching latest release tag from $REPO..."
tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
       | grep -oE '"tag_name":\s*"[^"]+' \
       | head -n1 \
       | sed -E 's/.*"([^"]+)$/\1/')"
if [[ -z "$tag" ]]; then
  echo "could not determine latest release tag" >&2
  exit 1
fi
version="${tag#v}"

# Skip work when the installed binary already advertises this version.
# main.version is wired in via -ldflags by goreleaser, so the binary
# self-reports through `-version` if available; otherwise we always
# update.
current="$( "$BIN_PATH" -version 2>/dev/null || true )"
if [[ -n "$current" && "$current" == *"$version"* ]]; then
  echo "already on $tag, nothing to do."
  exit 0
fi
echo "installing $tag (was: ${current:-unknown})"

asset="iwaki-ai_${version}_linux_${goarch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
echo "downloading $url..."
curl -fsSL -o "${tmp}/${asset}" "$url"
tar -xzf "${tmp}/${asset}" -C "$tmp"

install -m 0755 "${tmp}/iwaki-ai-bot" "$BIN_PATH"

if systemctl is-active --quiet "$SERVICE_NAME"; then
  systemctl restart "$SERVICE_NAME"
  echo "restarted ${SERVICE_NAME}"
else
  echo "${SERVICE_NAME} is not active; binary updated but service not restarted"
fi
