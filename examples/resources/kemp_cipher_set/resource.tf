resource "kemp_cipher_set" "modern_tls" {
  name = "ModernTLS"
  ciphers = [
    "ECDHE-ECDSA-AES256-GCM-SHA384",
    "ECDHE-RSA-AES256-GCM-SHA384",
    "ECDHE-ECDSA-CHACHA20-POLY1305",
    "ECDHE-RSA-CHACHA20-POLY1305",
  ]
}

# Assign the cipher set to a virtual service
resource "kemp_virtual_service" "example" {
  address    = "10.0.0.1"
  port       = "443"
  protocol   = "tcp"
  cipher_set = kemp_cipher_set.modern_tls.name
}
