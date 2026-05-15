terraform {
  required_version = ">= 1.6"

  required_providers {
    oci = {
      source  = "oracle/oci"
      version = "~> 6.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
    sops = {
      source  = "carlpett/sops"
      version = "~> 1.0"
    }
  }
}

# Decrypt secrets.enc.yaml using the operator's age key. SOPS picks
# the key up automatically from ~/.config/sops/age/keys.txt (or
# SOPS_AGE_KEY_FILE for CI). All sensitive material - OCI auth, bot
# tokens - lives inside this file; nothing secret should appear in
# *.tfvars or in process arguments.
data "sops_file" "secrets" {
  source_file = "${path.module}/secrets.enc.yaml"
}

provider "oci" {
  region           = data.sops_file.secrets.data["oci.region"]
  tenancy_ocid     = data.sops_file.secrets.data["oci.tenancy_ocid"]
  user_ocid        = data.sops_file.secrets.data["oci.user_ocid"]
  fingerprint      = data.sops_file.secrets.data["oci.fingerprint"]
  private_key_path = pathexpand(data.sops_file.secrets.data["oci.private_key_path"])
}
