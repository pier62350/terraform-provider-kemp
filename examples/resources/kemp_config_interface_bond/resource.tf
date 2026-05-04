resource "kemp_config_interface_bond" "bond0" {
  interface_id = "1"
  members      = ["2", "3"]
}
