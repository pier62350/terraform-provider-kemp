resource "kemp_match_content_rule" "is_admin_path" {
  name        = "is-admin-path"
  pattern     = "^/admin/"
  match_type  = "regex"
  ignore_case = true
}
