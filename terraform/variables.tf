# OCI authentication. All five fields are present in ~/.oci/config;
# pass them via terraform.tfvars (or env: TF_VAR_tenancy_ocid=...).
variable "tenancy_ocid" {
  description = "OCI tenancy OCID"
  type        = string
}

variable "user_ocid" {
  description = "OCI user OCID"
  type        = string
}

variable "fingerprint" {
  description = "OCI API key fingerprint"
  type        = string
}

variable "private_key_path" {
  description = "Path to the OCI API private key on the local machine"
  type        = string
  default     = "~/.oci/oci_api_key.pem"
}

variable "region" {
  description = "OCI region (e.g. ap-tokyo-1, ap-seoul-1)"
  type        = string
}

variable "config_file_profile" {
  description = "Profile inside ~/.oci/config; leave blank to use direct credential vars"
  type        = string
  default     = ""
}

# Where the resources live. Default to the root compartment (= tenancy)
# unless the user explicitly carves out a separate compartment.
variable "compartment_ocid" {
  description = "Compartment OCID to create resources in. Defaults to the tenancy (root compartment)."
  type        = string
  default     = ""
}

# Compute shape and sizing for the bot VM. A1.Flex is Always Free up
# to 4 OCPUs / 24 GB RAM aggregated across instances; one shape with
# 1 OCPU / 6 GB is plenty for a single Discord bot.
variable "instance_shape" {
  description = "OCI compute shape"
  type        = string
  default     = "VM.Standard.A1.Flex"
}

variable "instance_ocpus" {
  description = "Number of OCPUs (Flex shapes only)"
  type        = number
  default     = 1
}

variable "instance_memory_gb" {
  description = "Memory in GB (Flex shapes only)"
  type        = number
  default     = 6
}

variable "instance_name" {
  description = "Display name for the compute instance"
  type        = string
  default     = "iwaki-ai-bot"
}

# Where to drop the generated SSH key on the local filesystem so the
# operator can `ssh -i` straight away.
variable "ssh_private_key_path" {
  description = "Local path to write the generated SSH private key"
  type        = string
  default     = "~/.ssh/iwaki-ai"
}

# Restrict who can hit port 22 on the public VM. Defaults to anywhere
# so the operator can ssh from changing IPs, but locking this down to
# a specific CIDR is recommended once a stable address is known.
variable "ssh_ingress_cidr" {
  description = "CIDR allowed to SSH into the VM"
  type        = string
  default     = "0.0.0.0/0"
}

# --- Bot secrets passed into cloud-init ---------------------------
# These end up in /etc/iwaki-ai/iwaki-ai.env on the VM. Keep them out
# of git: set via terraform.tfvars (gitignored) or TF_VAR_* env vars.

variable "discord_bot_token" {
  description = "Discord Bot token (Bot prefix NOT included)"
  type        = string
  sensitive   = true
}

variable "discord_guild_id" {
  description = "Guild ID where the /summarize command is registered"
  type        = string
}

variable "gemini_api_key" {
  description = "Google AI Studio (Gemini) API key"
  type        = string
  sensitive   = true
}

variable "github_repo" {
  description = "GitHub repo (owner/name) that hosts releases and install.sh"
  type        = string
  default     = "masahiro331/iwaki-ai"
}
