resource "kemp_real_server" "example" {
  virtual_service_id = kemp_virtual_service.example.id
  address            = "10.0.0.10"
  port               = "8080"
  weight             = 1000
  forward            = "nat"
  enable             = true
}
