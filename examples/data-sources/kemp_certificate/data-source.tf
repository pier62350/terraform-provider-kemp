data "kemp_certificate" "existing" {
  name = "my-cert"
}

output "cert_type" {
  value = data.kemp_certificate.existing.type
}
