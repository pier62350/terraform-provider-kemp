resource "kemp_config_packet_routing_filter" "main" {
  enabled              = true
  drop                 = true
  restrict_to_interface = false
  include_wui          = false
}
