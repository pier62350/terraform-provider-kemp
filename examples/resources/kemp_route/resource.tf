resource "kemp_route" "example" {
  destination = "10.10.0.0/24"
  gateway     = "10.0.0.1"
}
