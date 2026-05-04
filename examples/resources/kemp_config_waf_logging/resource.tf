resource "kemp_config_waf_logging" "main" {
  enabled    = true
  remote_uri = "https://siem.example.com/waf"
  username   = "wafuser"
  password   = "secret"
  log_format = "cef"
}
