data "kemp_config_interface" "eth0" {
  id = 0
}

output "eth0_ip" {
  value = data.kemp_config_interface.eth0.ip_address
}
