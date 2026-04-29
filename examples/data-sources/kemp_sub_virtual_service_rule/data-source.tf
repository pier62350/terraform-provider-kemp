# Look up a rule attachment on a SubVS that was created outside Terraform.

data "kemp_sub_virtual_service_rule" "existing" {
  parent_virtual_service_id = "12"
  sub_virtual_service_id    = "15"
  rule                      = "match-api-path"
}
