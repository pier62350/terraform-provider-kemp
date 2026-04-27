resource "kemp_virtual_service" "example" {
  address  = "10.0.0.100"
  port     = "443"
  protocol = "tcp"
  type     = "http"
  nickname = "example-vs"
  enabled  = true
}
