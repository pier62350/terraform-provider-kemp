# Remote group with specific permissions
resource "kemp_config_group" "admins" {
  name        = "Domain Admins"
  permissions = ["real", "vs", "rules", "backup", "certs"]
}

# Read-only group
resource "kemp_config_group" "readonly" {
  name        = "Domain Users"
  permissions = []
}
