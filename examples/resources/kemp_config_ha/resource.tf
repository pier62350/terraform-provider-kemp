resource "kemp_config_ha" "main" {
  mode            = 1
  partner_address = "192.168.1.10"
  shared_address  = "192.168.1.1"
  secret          = var.ha_secret
}
