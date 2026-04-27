data "kemp_virtual_service_rule" "existing" {
  virtual_service_id = "12"
  direction          = "request"
  rule               = "match-api-path"
}
