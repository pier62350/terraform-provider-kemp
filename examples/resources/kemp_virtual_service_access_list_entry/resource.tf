resource "kemp_virtual_service" "web" {
  address  = "10.0.0.100"
  port     = "443"
  protocol = "tcp"
  type     = "http"
}

# Allow a specific IP on this virtual service.
resource "kemp_virtual_service_access_list_entry" "office" {
  virtual_service_id = kemp_virtual_service.web.id
  list_type          = "allow"
  address            = "10.0.0.5"
}

# Block a specific IP on this virtual service.
resource "kemp_virtual_service_access_list_entry" "blocked" {
  virtual_service_id = kemp_virtual_service.web.id
  list_type          = "block"
  address            = "203.0.113.99"
}
