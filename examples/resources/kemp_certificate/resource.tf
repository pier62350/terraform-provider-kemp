resource "kemp_certificate" "example" {
  name = "wildcard-example-com"

  # Read a PEM bundle (cert + key concatenated) and base64-encode it.
  data = base64encode(file("${path.module}/wildcard-example-com.pem"))
}

# For a PFX with a passphrase:
#
# resource "kemp_certificate" "pfx" {
#   name     = "imported-pfx"
#   data     = base64encode(file("${path.module}/cert.pfx"))
#   password = var.pfx_password
# }
