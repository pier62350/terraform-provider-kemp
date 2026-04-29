resource "kemp_config_ldap_endpoint" "corp" {
  name      = "CORP_LDAP"
  ldap_type = "starttls"
  server    = "ldap1.corp.example.com ldap2.corp.example.com"

  revalidation_interval = 120
  referral_count        = 2
}
