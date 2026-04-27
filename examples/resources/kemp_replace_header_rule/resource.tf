resource "kemp_replace_header_rule" "rewrite_host" {
  name        = "rewrite-host-header"
  header      = "Host"
  pattern     = "^public\\.example\\.com$"
  replacement = "internal.example.local"
}
