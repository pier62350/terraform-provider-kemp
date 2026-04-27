resource "kemp_owasp_custom_rule" "company_rules" {
  filename = "company_owasp.conf"
  data     = base64encode(file("${path.module}/company_owasp.conf"))
}
