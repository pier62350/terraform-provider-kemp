resource "kemp_gslb_cluster" "example" {
  ip   = "10.0.0.5"
  name = "MyCluster"
  type = "remoteLM"

  lat_secs = 3661
  lon_secs = 3661
}
