data "kemp_acme_certificate" "existing" {
  name      = "example.com"
  acme_type = "1"
}

output "expiry_date" {
  value = data.kemp_acme_certificate.existing.expiry_date
}
