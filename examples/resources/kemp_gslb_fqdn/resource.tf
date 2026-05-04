resource "kemp_gslb_cluster" "primary" {
  ip   = "10.0.0.5"
  name = "PrimaryCluster"
  type = "remoteLM"
}

resource "kemp_gslb_fqdn" "example" {
  fqdn               = "example.com"
  selection_criteria = "wrr"
  fail_time          = 1000

  member {
    ip      = "10.0.0.1"
    cluster = kemp_gslb_cluster.primary.name
    checker = "tcp"
  }

  member {
    ip = "10.0.0.2"
  }
}
