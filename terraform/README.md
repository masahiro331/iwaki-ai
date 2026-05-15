# Terraform: Iwaki AI bot on Oracle Cloud (Always Free)

Stands up a single Ampere A1.Flex VM in your OCI tenancy, generates an
SSH keypair on the fly, opens port 22 to the world (by default), and
cloud-inits the bot so that it self-installs, drops the env file, and
enables the systemd unit.

Bot/AI secrets (Discord token, Gemini key) live in `secrets.enc.yaml`,
which is SOPS-encrypted with age so the file is safe to commit. OCI
credentials are read from `~/.oci/config` and shared with the OCI CLI
rather than duplicated here.

## Prereqs

- An OCI tenancy with Always Free A1 capacity.
- `~/.oci/config` set up (the OCI CLI quickstart covers this) and the
  matching API key registered on your user.
- `brew install sops age`
- A Discord bot token, target guild ID, and a Gemini API key.

## One-time setup (per operator)

```bash
# 1) Make an age key on this machine
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
# note the "Public key: age1..." line

# 2) Register it as a recipient in .sops.yaml
# Edit terraform/.sops.yaml and replace the placeholder with your
# age1... public key. Existing operators rotate via `sops updatekeys`.

# 3) Author the secrets file
cd terraform
cp secrets.example.yaml secrets.yaml
$EDITOR secrets.yaml        # fill in discord/gemini values
sops --encrypt secrets.yaml > secrets.enc.yaml
rm secrets.yaml             # keep only the encrypted copy

# Subsequent edits:
sops secrets.enc.yaml       # opens $EDITOR with the decrypted content
```

## Apply

`tenancy_ocid` is the only OCI value Terraform needs explicitly (used
as the default compartment for resources). Grab it from
`~/.oci/config` or the OCI console and pass it as a var:

```bash
terraform init
terraform plan  -var tenancy_ocid=ocid1.tenancy.oc1..xxx
terraform apply -var tenancy_ocid=ocid1.tenancy.oc1..xxx
```

Or stash it in `terraform.tfvars` (gitignored):

```bash
echo 'tenancy_ocid = "ocid1.tenancy.oc1..xxx"' > terraform.tfvars
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
sudo /tmp/install.sh   # idempotent; or use scripts/update.sh
```

`update.sh` is idempotent and only restarts the service when the
binary actually changed.

## Tear down

```bash
terraform destroy
```

Removes the VM, networking, and the local SSH key. The OCI compartment
and the encrypted secrets file are left alone.
