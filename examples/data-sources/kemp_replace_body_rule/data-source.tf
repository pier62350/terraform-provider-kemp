data "kemp_replace_body_rule" "existing" {
  name = "sanitize-response"
}

output "pattern" {
  value = data.kemp_replace_body_rule.existing.pattern
}
