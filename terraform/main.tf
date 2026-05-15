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
# SOPS_AGE_KEY_FILE for CI). Only Discord/Gemini secrets live here;
# OCI authentication is read from ~/.oci/config so it can be shared
# with the OCI CLI without going through SOPS.
data "sops_file" "secrets" {
  source_file = "${path.module}/secrets.enc.yaml"
}

provider "oci" {
  config_file_profile = var.oci_config_profile
}
