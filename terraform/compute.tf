# --- SSH key -----------------------------------------------------
# Generate a fresh ed25519 keypair so the operator never has to
# decide which existing key to reuse, and write the private half to
# disk with 0600 so `ssh -i` works without further chmod.

resource "tls_private_key" "ssh" {
  algorithm = "ED25519"
}

resource "local_sensitive_file" "ssh_private_key" {
  content         = tls_private_key.ssh.private_key_openssh
  filename        = pathexpand(var.ssh_private_key_path)
  file_permission = "0600"
}

resource "local_file" "ssh_public_key" {
  content         = tls_private_key.ssh.public_key_openssh
  filename        = pathexpand("${var.ssh_private_key_path}.pub")
  file_permission = "0644"
}

# --- Ubuntu image ------------------------------------------------
# Pick the most recent Canonical Ubuntu 24.04 ARM64 image OCI publishes
# for this region. Filtering by operating system + version + shape keeps
# us off x86 images when running on A1.Flex.

data "oci_core_images" "ubuntu" {
  compartment_id           = local.compartment_ocid
  operating_system         = "Canonical Ubuntu"
  operating_system_version = "24.04"
  shape                    = var.instance_shape
  sort_by                  = "TIMECREATED"
  sort_order               = "DESC"
}

data "oci_identity_availability_domains" "ads" {
  compartment_id = local.compartment_ocid
}

# --- Compute instance -------------------------------------------

resource "oci_core_instance" "bot" {
  compartment_id      = local.compartment_ocid
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  display_name        = var.instance_name
  shape               = var.instance_shape

  shape_config {
    ocpus         = var.instance_ocpus
    memory_in_gbs = var.instance_memory_gb
  }

  source_details {
    source_type = "image"
    source_id   = data.oci_core_images.ubuntu.images[0].id
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.public.id
    assign_public_ip = true
    hostname_label   = "iwakiaibot"
  }

  metadata = {
    ssh_authorized_keys = tls_private_key.ssh.public_key_openssh
    # cloud-init must be base64-encoded for OCI's user_data field.
    # Secret values come straight from SOPS so they never appear in
    # plaintext on disk under terraform/.
    user_data = base64encode(templatefile("${path.module}/cloud-init.yaml.tpl", {
      discord_bot_token = data.sops_file.secrets.data["discord.bot_token"]
      discord_guild_id  = data.sops_file.secrets.data["discord.guild_id"]
      gemini_api_key    = data.sops_file.secrets.data["gemini.api_key"]
      repo              = var.github_repo
    }))
  }
}
