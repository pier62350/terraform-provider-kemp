data "kemp_match_content_rule" "existing" {
  name = "match-api-path"
}

output "pattern" {
  value = data.kemp_match_content_rule.existing.pattern
}
