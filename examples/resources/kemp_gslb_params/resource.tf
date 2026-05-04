resource "kemp_gslb_params" "main" {
  zone               = "example.com"
  source_of_authority = "ns1.example.com"
  nameserver         = "ns1.example.com"
  soa_email          = "admin@example.com"
  ttl                = 3600
  persist            = 60000
}
