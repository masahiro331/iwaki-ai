#cloud-config
# cloud-init that brings a fresh OCI A1.Flex instance up to a runnable
# iwaki-ai bot. Order of operations:
#   1. write_files - drop the env file + sshd hardening on disk
#   2. runcmd      - fetch install.sh, swap ownership to iwaki-ai,
#                    start the service, scrub user-data
#
# Anything that has to run after the iwaki-ai system user exists lives
# in runcmd because install.sh is what creates that user.

hostname: iwaki-ai-bot
timezone: Asia/Tokyo

package_update: true
package_upgrade: false
packages:
  - curl
  - ca-certificates
  - tar
  - unattended-upgrades

write_files:
  # The env file is owned root:root for now; ownership is rewritten to
  # root:iwaki-ai inside runcmd once useradd has been called. cloud-init
  # creates /etc/iwaki-ai/ implicitly when this file is written, so we
  # don't need a separate mkdir.
  - path: /etc/iwaki-ai/iwaki-ai.env
    permissions: '0640'
    owner: root:root
    content: |
      DISCORD_BOT_TOKEN=${discord_bot_token}
      DISCORD_GUILD_ID=${discord_guild_id}
      GEMINI_API_KEY=${gemini_api_key}

  # Enable unattended security updates. The default Ubuntu config
  # already covers security pockets; this snippet just turns on the
  # nightly run and reboots at 04:00 JST if the kernel changed so
  # the bot doesn't drift behind on CVEs.
  - path: /etc/apt/apt.conf.d/20auto-upgrades
    permissions: '0644'
    owner: root:root
    content: |
      APT::Periodic::Update-Package-Lists "1";
      APT::Periodic::Unattended-Upgrade "1";

  - path: /etc/apt/apt.conf.d/52unattended-upgrades-reboot
    permissions: '0644'
    owner: root:root
    content: |
      Unattended-Upgrade::Automatic-Reboot "true";
      Unattended-Upgrade::Automatic-Reboot-Time "04:00";

  # Make sshd's policy explicit. Stock Ubuntu cloud images already
  # ship with PasswordAuthentication no, but a drop-in here survives
  # base image upgrades and documents intent.
  - path: /etc/ssh/sshd_config.d/10-iwaki-ai-hardening.conf
    permissions: '0644'
    owner: root:root
    content: |
      PasswordAuthentication no
      ChallengeResponseAuthentication no
      KbdInteractiveAuthentication no
      PermitRootLogin no
      PubkeyAuthentication yes

  # Single-script bootstrap. Using a script (instead of more runcmd
  # lines) lets us turn on `set -euo pipefail` so the box fails to
  # provision loudly rather than silently skipping a step.
  - path: /var/lib/iwaki-ai/bootstrap.sh
    permissions: '0700'
    owner: root:root
    content: |
      #!/usr/bin/env bash
      set -euo pipefail

      REPO="${repo}"
      LOG="/var/log/iwaki-ai-bootstrap.log"
      exec > >(tee -a "$LOG") 2>&1

      echo "[$(date -Is)] fetching install.sh from $REPO"
      curl -fsSL -o /var/tmp/install.sh \
        "https://raw.githubusercontent.com/$REPO/main/scripts/install.sh"
      chmod +x /var/tmp/install.sh
      /var/tmp/install.sh "$REPO"

      # install.sh seeds /etc/iwaki-ai/ as root:iwaki-ai 0750 only when
      # it didn't already exist; cloud-init created it first, so tighten
      # ownership and permissions here to match the systemd unit's
      # expectations.
      chown -R root:iwaki-ai /etc/iwaki-ai
      chmod 0750 /etc/iwaki-ai
      chmod 0640 /etc/iwaki-ai/iwaki-ai.env

      systemctl enable --now iwaki-ai-bot.service
      systemctl reload ssh || systemctl reload sshd

      # Health check. The bot logs in to Discord and registers the
      # slash command at boot, so give it ~15s before declaring
      # failure. Dump the journal on failure so the cloud-init log
      # captures *why* it didn't come up (bad token, missing env,
      # binary panic, etc).
      for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
        if systemctl is-active --quiet iwaki-ai-bot.service; then
          break
        fi
        sleep 1
      done
      if ! systemctl is-active --quiet iwaki-ai-bot.service; then
        echo "[$(date -Is)] iwaki-ai-bot.service failed to come up"
        journalctl -u iwaki-ai-bot.service --no-pager -n 50 || true
        exit 1
      fi

      # We can't remove the cloud-init user_data from the OCI metadata
      # service - it's a read-only OCI API. Anyone with shell access
      # to this VM can still curl 169.254.169.254 and base64-decode
      # the secrets. SSH is pubkey-only and the iwaki-ai user is
      # nologin, so practical exposure equals "whoever has the ubuntu
      # operator key", which is the same surface as /etc/iwaki-ai.

      rm -f /var/tmp/install.sh
      echo "[$(date -Is)] bootstrap complete"

runcmd:
  - [ /var/lib/iwaki-ai/bootstrap.sh ]
