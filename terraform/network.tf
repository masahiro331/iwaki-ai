# All networking the bot needs is outbound HTTPS (Discord gateway and
# Gemini) and a small inbound SSH window for operator access. Nothing
# inside the VCN talks to anything else, so a single public subnet
# with a permissive egress rule and a tight ingress rule is enough.

locals {
  # Fall back to the tenancy (root compartment) when no explicit
  # compartment is configured.
  compartment_ocid = var.compartment_ocid != "" ? var.compartment_ocid : var.tenancy_ocid
}

resource "oci_core_vcn" "main" {
  compartment_id = local.compartment_ocid
  cidr_blocks    = ["10.0.0.0/16"]
  display_name   = "iwaki-ai-vcn"
  dns_label      = "iwakiai"
}

resource "oci_core_internet_gateway" "main" {
  compartment_id = local.compartment_ocid
  vcn_id         = oci_core_vcn.main.id
  display_name   = "iwaki-ai-igw"
  enabled        = true
}

resource "oci_core_route_table" "public" {
  compartment_id = local.compartment_ocid
  vcn_id         = oci_core_vcn.main.id
  display_name   = "iwaki-ai-rt-public"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.main.id
  }
}

resource "oci_core_security_list" "public" {
  compartment_id = local.compartment_ocid
  vcn_id         = oci_core_vcn.main.id
  display_name   = "iwaki-ai-sl-public"

  # Permit all outbound traffic; the bot needs Discord (443) and
  # Gemini (443), and locking down by destination IP would force
  # tracking Discord/Gemini ranges over time.
  egress_security_rules {
    protocol         = "all"
    destination      = "0.0.0.0/0"
    destination_type = "CIDR_BLOCK"
    stateless        = false
  }

  # SSH ingress for the operator only. Tighten to a single source CIDR
  # by overriding `ssh_ingress_cidr` if a fixed home/office IP is
  # available.
  ingress_security_rules {
    protocol  = "6" # TCP
    source    = var.ssh_ingress_cidr
    stateless = false

    tcp_options {
      min = 22
      max = 22
    }
  }
}

resource "oci_core_subnet" "public" {
  compartment_id             = local.compartment_ocid
  vcn_id                     = oci_core_vcn.main.id
  cidr_block                 = "10.0.1.0/24"
  display_name               = "iwaki-ai-subnet-public"
  dns_label                  = "public"
  route_table_id             = oci_core_route_table.public.id
  security_list_ids          = [oci_core_security_list.public.id]
  prohibit_public_ip_on_vnic = false
}
