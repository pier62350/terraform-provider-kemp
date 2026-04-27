data "kemp_replace_header_rule" "existing" {
  name = "rewrite-location"
}

output "replacement" {
  value = data.kemp_replace_header_rule.existing.replacement
}
