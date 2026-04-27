resource "kemp_delete_header_rule" "strip_server" {
  name    = "strip-server-header"
  pattern = "^Server$"
}
