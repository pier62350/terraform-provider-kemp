data "kemp_add_header_rule" "existing" {
  name = "add-x-forwarded-for"
}

output "header" {
  value = data.kemp_add_header_rule.existing.header
}
