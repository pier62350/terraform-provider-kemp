# Add an IP to the appliance-wide allow list.
resource "kemp_config_access_list_entry" "office_allow" {
  list_type = "allow"
  address   = "10.0.0.0/24"
}

# Add an IP to the appliance-wide block list.
resource "kemp_config_access_list_entry" "bad_actor_block" {
  list_type = "block"
  address   = "203.0.113.42"
}
