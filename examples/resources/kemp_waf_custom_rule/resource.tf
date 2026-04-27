resource "kemp_waf_custom_rule" "brute_force" {
  filename = "modsecurity_crs_11_brute_force"
  data     = base64encode(file("${path.module}/modsecurity_crs_11_brute_force"))
}
