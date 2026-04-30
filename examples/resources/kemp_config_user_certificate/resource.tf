# Generate a client certificate for an admin user
resource "kemp_config_user_certificate" "admin" {
  username   = "admin"
  passphrase = var.cert_passphrase
}

output "admin_certificate" {
  value     = kemp_config_user_certificate.admin.certificate
  sensitive = true
}
