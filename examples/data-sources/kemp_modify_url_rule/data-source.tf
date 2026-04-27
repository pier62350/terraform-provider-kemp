data "kemp_modify_url_rule" "existing" {
  name = "strip-api-prefix"
}

output "pattern" {
  value = data.kemp_modify_url_rule.existing.pattern
}
