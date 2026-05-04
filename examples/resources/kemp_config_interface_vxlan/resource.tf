resource "kemp_config_interface_vxlan" "tunnel1" {
  interface_id = "1"
  vni          = 1000
  remote       = "10.0.0.5"
}
