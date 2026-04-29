resource "kemp_local_user" "ops" {
  username = "ops01"
  password = var.ops_password

  permissions = ["vs", "real"]
}

# Read-only user (no permissions)
resource "kemp_local_user" "readonly" {
  username = "monitor01"
  password = var.monitor_password
}
