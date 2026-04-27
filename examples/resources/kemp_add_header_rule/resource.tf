resource "kemp_add_header_rule" "x_forwarded_proto" {
  name        = "x-forwarded-proto-https"
  header      = "X-Forwarded-Proto"
  replacement = "https"
}
