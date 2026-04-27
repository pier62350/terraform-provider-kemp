resource "kemp_modify_url_rule" "strip_v1_prefix" {
  name        = "strip-api-v1-prefix"
  pattern     = "^/api/v1/(.*)$"
  replacement = "/$1"
}
