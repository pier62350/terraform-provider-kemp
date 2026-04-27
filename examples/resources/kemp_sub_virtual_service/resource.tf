resource "kemp_virtual_service" "frontend" {
  address          = "10.0.0.100"
  port             = "443"
  protocol         = "tcp"
  type             = "http"
  ssl_acceleration = true
  cert_files       = [kemp_certificate.wildcard.name]
}

resource "kemp_certificate" "wildcard" {
  name = "wildcard"
  data = base64encode(file("${path.module}/cert.pem"))
}

resource "kemp_sub_virtual_service" "api" {
  parent_id = kemp_virtual_service.frontend.id
  nickname  = "api-tier"
  enabled   = true

  esp_enabled         = true
  esp_input_auth_mode = "form"
}

resource "kemp_sub_virtual_service" "static" {
  parent_id = kemp_virtual_service.frontend.id
  nickname  = "static-assets"
  enabled   = true
}
