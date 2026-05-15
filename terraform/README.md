# Terraform: Iwaki AI bot on Oracle Cloud (Always Free)

Stands up a single Ampere A1.Flex VM in your OCI tenancy, generates an
SSH keypair on the fly, opens port 22 to the world (by default), and
cloud-inits the bot so that it self-installs, drops the env file, and
enables the systemd unit.

## Prereqs

- An OCI tenancy with the Always Free A1 capacity (1 OCPU / 6 GB is
  enough; raise `instance_*` to grow inside the free aggregate of 4
  OCPUs / 24 GB).
- An OCI API key registered on your user; the five values from your
  `~/.oci/config` need to land in `terraform.tfvars`.
- A Discord bot token, target guild ID, and a Gemini API key.

## Use

```bash
cp terraform.tfvars.example terraform.tfvars
$EDITOR terraform.tfvars   # fill in OCI auth + bot secrets

terraform init
terraform plan
terraform apply
```

Outputs the public IP and an `ssh -i ...` command using the generated
key written to `~/.ssh/iwaki-ai` by default.

cloud-init pulls `scripts/install.sh` from `main` on GitHub, downloads
the latest release archive for the host arch, installs the binary at
`/usr/local/bin/iwaki-ai-bot`, drops the env file, and starts
`iwaki-ai-bot.service`. Give it a couple of minutes after `apply`
returns before expecting `/summarize` to respond.

## Updating the bot

Tag a new release in the repo (`git tag v0.2.0 && git push origin v0.2.0`)
and the GoReleaser workflow publishes the binaries. Then on the VM:

```bash
ssh -i ~/.ssh/iwaki-ai ubuntu@<public_ip>
sudo /tmp/install.sh   # or: curl -fsSL .../scripts/update.sh | sudo bash
```

`update.sh` is idempotent and only restarts the service when the
binary actually changed.

## Tear down

```bash
terraform destroy
```

Removes the VM, networking, and the local SSH key. The OCI compartment
itself is left alone.
