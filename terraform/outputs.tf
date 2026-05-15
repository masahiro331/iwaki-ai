output "instance_public_ip" {
  description = "Public IPv4 address of the bot VM"
  value       = oci_core_instance.bot.public_ip
}

output "ssh_command" {
  description = "Ready-to-paste ssh command using the generated key"
  value       = "ssh -i ${local_sensitive_file.ssh_private_key.filename} ubuntu@${oci_core_instance.bot.public_ip}"
}

output "ssh_private_key_path" {
  description = "Filesystem location of the generated SSH private key"
  value       = local_sensitive_file.ssh_private_key.filename
}
