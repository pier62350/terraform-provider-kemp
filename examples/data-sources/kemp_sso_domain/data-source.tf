# Read an existing SSO domain to reference its settings
data "kemp_sso_domain" "corp" {
  name = "corp.example.com"
}

output "corp_auth_type" {
  value = data.kemp_sso_domain.corp.auth_type
}
