resource "kemp_config_waf_update" "main" {
  auto_update  = true
  auto_install = true
  install_hour = 3
}
