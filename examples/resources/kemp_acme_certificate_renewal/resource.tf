# Force a manual renewal. Change the `version` trigger value to re-fire
# renewacmecert on the next `terraform apply`.

resource "kemp_acme_certificate_renewal" "force" {
  cert_name = kemp_acme_certificate.my_cert.name
  acme_type = "1"

  triggers = {
    version = "1"
  }
}
