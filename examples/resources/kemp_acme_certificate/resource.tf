resource "kemp_virtual_service" "http_challenge" {
  address  = "10.0.0.100"
  port     = "80"
  protocol = "tcp"
  type     = "http"
  nickname = "le-challenge"
  enabled  = true
}

resource "kemp_acme_certificate" "example" {
  name               = "example-le"
  common_name        = "www.example.com"
  virtual_service_id = kemp_virtual_service.http_challenge.id
  acme_type          = "letsencrypt"
  key_size           = 2048
  email              = "ops@example.com"
}

# Wildcard via DNS-01:
# resource "kemp_acme_certificate" "wildcard" {
#   name               = "example-wild"
#   common_name        = "*.example.com"
#   virtual_service_id = kemp_virtual_service.http_challenge.id
#   acme_type          = "letsencrypt"
#   dns_api            = "godaddy.com"
#   dns_api_params     = var.godaddy_credentials
# }
