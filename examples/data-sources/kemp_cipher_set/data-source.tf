# Read a built-in cipher set
data "kemp_cipher_set" "best_practices" {
  name = "BestPractices"
}

# Read a custom cipher set managed elsewhere
data "kemp_cipher_set" "custom" {
  name = "ModernTLS"
}

output "best_practices_ciphers" {
  value = data.kemp_cipher_set.best_practices.ciphers
}
