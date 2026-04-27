# Let's Encrypt account — one-time bootstrap, required before any
# kemp_acme_certificate with acme_type = "1" can be issued.
resource "kemp_acme_account" "letsencrypt" {
  acme_type     = "1"
  email         = "ops@example.com"
  directory_url = "https://acme-staging-v02.api.letsencrypt.org/directory"
  renew_period  = 30
}

# DigiCert account — supply your EAB Key ID and HMAC.
resource "kemp_acme_account" "digicert" {
  acme_type = "2"
  kid       = var.digicert_kid
  hmac_key  = var.digicert_hmac
}
