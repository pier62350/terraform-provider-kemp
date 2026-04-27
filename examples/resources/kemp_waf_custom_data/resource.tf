resource "kemp_waf_custom_data" "bad_robots" {
  filename = "modsecurity_35_bad_robots.data"
  data     = base64encode(file("${path.module}/modsecurity_35_bad_robots.data"))
}
