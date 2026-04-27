data "kemp_delete_header_rule" "existing" {
  name = "remove-server-header"
}

output "pattern" {
  value = data.kemp_delete_header_rule.existing.pattern
}
