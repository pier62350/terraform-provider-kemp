data "kemp_certificates" "all" {}

output "cert_names" {
  value = [for c in data.kemp_certificates.all.certificates : c.name]
}
