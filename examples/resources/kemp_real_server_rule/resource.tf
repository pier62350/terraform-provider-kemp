resource "kemp_virtual_service" "web" {
  address  = "10.0.0.100"
  port     = "80"
  protocol = "tcp"
  type     = "http"
}

resource "kemp_real_server" "backend" {
  virtual_service_id = kemp_virtual_service.web.id
  address            = "192.168.1.10"
  port               = "8080"
  weight             = 100
}

resource "kemp_match_content_rule" "health" {
  name    = "health-check"
  pattern = "/health"
}

# Attach the rule to the real server.
resource "kemp_real_server_rule" "backend_health" {
  virtual_service_id = kemp_virtual_service.web.id
  real_server_id     = kemp_real_server.backend.id
  vs_port            = kemp_virtual_service.web.port
  vs_protocol        = kemp_virtual_service.web.protocol
  rs_address         = kemp_real_server.backend.address
  rs_port            = kemp_real_server.backend.port
  rule               = kemp_match_content_rule.health.name
}
