resource "kemp_virtual_service" "example" {
  address  = "192.168.1.200"
  port     = "443"
  protocol = "tcp"
  type     = "http"
  nickname = "example-vs"
  enabled  = true
}
