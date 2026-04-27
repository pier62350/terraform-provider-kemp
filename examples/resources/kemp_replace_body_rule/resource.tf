resource "kemp_replace_body_rule" "redact_internal_hostname" {
  name        = "redact-internal-hostname"
  pattern     = "internal\\.example\\.local"
  replacement = "[redacted]"
}
