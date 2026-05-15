#!/usr/bin/env bash
# install.sh - one-shot bootstrap for a fresh VM.
#
# Creates the iwaki-ai system user, fetches the latest release binary
# from GitHub, installs the systemd unit, and prepares the env file
# (left empty so secrets can be filled in afterwards). Idempotent.
#
# Usage:
#   sudo ./install.sh [REPO]
# where REPO defaults to masahiro331/iwaki-ai.
set -euo pipefail

REPO="${1:-masahiro331/iwaki-ai}"
BIN_PATH="/usr/local/bin/iwaki-ai-bot"
ENV_DIR="/etc/iwaki-ai"
ENV_FILE="${ENV_DIR}/iwaki-ai.env"
SERVICE_NAME="iwaki-ai-bot.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"

if [[ $EUID -ne 0 ]]; then
  echo "install.sh must be run as root (use sudo)" >&2
  exit 1
fi

# 1) Detect host architecture and map it to GoReleaser's archive naming.
arch="$(uname -m)"
case "$arch" in
  x86_64)  goarch="amd64" ;;
  aarch64) goarch="arm64" ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

# 2) Create the unprivileged service user once.
if ! id iwaki-ai >/dev/null 2>&1; then
  useradd --system --no-create-home --shell /usr/sbin/nologin iwaki-ai
fi

# 3) Pull the latest release tag from GitHub.
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
echo "latest tag: $tag (version $version)"

# 4) Download the matching archive into a scratch dir.
asset="iwaki-ai_${version}_linux_${goarch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
echo "downloading $url..."
curl -fsSL -o "${tmp}/${asset}" "$url"
tar -xzf "${tmp}/${asset}" -C "$tmp"

# 5) Install the bot binary atomically and set permissions.
install -m 0755 "${tmp}/iwaki-ai-bot" "$BIN_PATH"

# 6) Prepare the env file the first time only. Leave existing values intact.
install -d -o root -g iwaki-ai -m 0750 "$ENV_DIR"
if [[ ! -f "$ENV_FILE" ]]; then
  cat > "$ENV_FILE" <<'EOF'
DISCORD_BOT_TOKEN=
DISCORD_GUILD_ID=
GEMINI_API_KEY=
EOF
  chown root:iwaki-ai "$ENV_FILE"
  chmod 0640 "$ENV_FILE"
  echo "created blank ${ENV_FILE}; edit it to add real secrets before starting the service"
fi

# 7) Install/refresh the systemd unit.
script_dir="$(cd "$(dirname "$0")" && pwd)"
unit_src="${script_dir}/../deploy/systemd/${SERVICE_NAME}"
if [[ ! -f "$unit_src" ]]; then
  echo "systemd unit not found at $unit_src" >&2
  exit 1
fi
install -m 0644 "$unit_src" "$SERVICE_PATH"
systemctl daemon-reload

echo "install complete."
echo "next steps:"
echo "  1. sudo \$EDITOR ${ENV_FILE}  # fill in DISCORD_BOT_TOKEN, DISCORD_GUILD_ID, GEMINI_API_KEY"
echo "  2. sudo systemctl enable --now ${SERVICE_NAME}"
echo "  3. sudo systemctl status ${SERVICE_NAME}"
