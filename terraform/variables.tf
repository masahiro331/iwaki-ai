# Operator-facing knobs only. Bot/AI secrets live in secrets.enc.yaml
# (SOPS+age) and are read through data.sops_file.secrets in main.tf.
# OCI credentials are sourced from ~/.oci/config so the same profile
# works for the OCI CLI and Terraform.

variable "oci_config_profile" {
  description = "Profile in ~/.oci/config to use for authentication."
  type        = string
  default     = "DEFAULT"
}

variable "tenancy_ocid" {
  description = "OCI tenancy OCID. Used as the default compartment when compartment_ocid is empty. Lookup via `oci iam compartment list` or the OCI console."
  type        = string
}

variable "compartment_ocid" {
  description = "Compartment OCID. Defaults to the tenancy (root compartment)."
  type        = string
  default     = ""
}

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

variable "ssh_private_key_path" {
  description = "Local path to write the generated SSH private key"
  type        = string
  default     = "~/.ssh/iwaki-ai"
}

variable "ssh_ingress_cidr" {
  description = "CIDR allowed to SSH into the VM"
  type        = string
  default     = "0.0.0.0/0"
}

variable "github_repo" {
  description = "GitHub repo (owner/name) that hosts releases and install.sh"
  type        = string
  default     = "masahiro331/iwaki-ai"
}
