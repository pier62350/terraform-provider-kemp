# One-time ACME service bootstrap. Required before any kemp_acme_certificate
# can be issued.

resource "kemp_acme_account" "letsencrypt" {
  acme_type     = "1" # 1 = Let's Encrypt, 2 = DigiCert
  email         = "ops@example.com"
  directory_url = "https://acme-staging-v02.api.letsencrypt.org/directory"
  renew_period  = 30
}
