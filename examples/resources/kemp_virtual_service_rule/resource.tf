resource "kemp_virtual_service" "web" {
  address  = "10.0.0.100"
  port     = "443"
  protocol = "tcp"
  type     = "http"
}

resource "kemp_add_header_rule" "x_forwarded_proto" {
  name        = "x-forwarded-proto-https"
  header      = "X-Forwarded-Proto"
  replacement = "https"
}

# Attach the rule to the VS in the request direction.
resource "kemp_virtual_service_rule" "web_xfwd_proto" {
  virtual_service_id = kemp_virtual_service.web.id
  rule               = kemp_add_header_rule.x_forwarded_proto.name
  direction          = "request" # request | response | responsebody | pre
}
