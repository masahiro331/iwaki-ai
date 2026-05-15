#cloud-config
# cloud-init that brings a fresh OCI A1.Flex instance up to a runnable
# iwaki-ai bot. Reads the env values that were injected by Terraform
# and lays them down at /etc/iwaki-ai/iwaki-ai.env, then runs the
# repo's install.sh to fetch the binary and register the systemd unit.

package_update: true
package_upgrade: false
packages:
  - curl
  - ca-certificates
  - tar

write_files:
  - path: /etc/iwaki-ai/iwaki-ai.env
    permissions: '0640'
    owner: root:root  # ownership is fixed in runcmd below after the user exists
    content: |
      DISCORD_BOT_TOKEN=${discord_bot_token}
      DISCORD_GUILD_ID=${discord_guild_id}
      GEMINI_API_KEY=${gemini_api_key}

  # Force pubkey-only SSH. Cloud images already lean this way, but a
  # drop-in under sshd_config.d makes it explicit and survives base
  # image changes.
  - path: /etc/ssh/sshd_config.d/10-iwaki-ai-hardening.conf
    permissions: '0644'
    owner: root:root
    content: |
      PasswordAuthentication no
      ChallengeResponseAuthentication no
      KbdInteractiveAuthentication no
      PermitRootLogin no
      PubkeyAuthentication yes

runcmd:
  # Fetch the install.sh that lives in the repo. Using the script
  # directly off main avoids having to package it into the release
  # archive, and it's idempotent.
  - curl -fsSL -o /tmp/install.sh https://raw.githubusercontent.com/${repo}/main/scripts/install.sh
  - chmod +x /tmp/install.sh
  - /tmp/install.sh ${repo}
  # install.sh writes a blank env file the first time, but we already
  # wrote the populated one above. Restore ownership/perms in case
  # the script touched them, then enable & start the service.
  - chown root:iwaki-ai /etc/iwaki-ai/iwaki-ai.env
  - chmod 0640 /etc/iwaki-ai/iwaki-ai.env
  - systemctl enable --now iwaki-ai-bot.service
  # Apply the hardened sshd drop-in. Existing sessions stay alive.
  - systemctl reload ssh || systemctl reload sshd
