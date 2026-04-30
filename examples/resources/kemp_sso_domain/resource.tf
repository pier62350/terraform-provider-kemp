# LDAP domain using a managed LDAP endpoint
resource "kemp_sso_domain" "ldap" {
  name          = "corp.example.com"
  auth_type     = "LDAP-StartTLS"
  ldap_endpoint = kemp_config_ldap_endpoint.corp.name
  logon_fmt     = "Principalname"
  logon_domain  = "CORP"

  max_failed_auths  = 5
  reset_fail_tout   = 300
  unblock_tout      = 1800
  sess_tout_idle_pub  = 900
  sess_tout_duration_pub  = 1800
  sess_tout_idle_priv = 3600
  sess_tout_duration_priv = 28800
}

# RADIUS domain
resource "kemp_sso_domain" "radius" {
  name                 = "radius.example.com"
  auth_type            = "RADIUS"
  server               = "192.168.1.10"
  radius_shared_secret = var.radius_secret
  radius_send_nas_id   = true
  logon_fmt            = "Username only"
}

# OIDC / OAuth domain (Azure AD / Entra ID)
resource "kemp_sso_domain" "azure_ad" {
  name              = "azure.example.com"
  auth_type         = "OIDC-OAUTH"
  oidc_app_id       = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  oidc_auth_ep_url  = "https://login.microsoftonline.com/TENANT_ID/oauth2/v2.0/authorize"
  oidc_token_ep_url = "https://login.microsoftonline.com/TENANT_ID/oauth2/v2.0/token"
  oidc_redirect_uri = "https://app.example.com/oidcauthz"
  oidc_logoff_url   = "https://login.microsoftonline.com/TENANT_ID/oauth2/v2.0/logout"
  oidc_secret       = var.oidc_secret
}

# SAML domain
resource "kemp_sso_domain" "saml" {
  name          = "saml.example.com"
  auth_type     = "LDAP-StartTLS"
  ldap_endpoint = kemp_config_ldap_endpoint.corp.name
  idp_entity_id = "https://sts.windows.net/TENANT_ID/"
  idp_sso_url   = "https://login.microsoftonline.com/TENANT_ID/saml2"
  idp_logoff_url = "https://login.microsoftonline.com/TENANT_ID/saml2"
  idp_cert      = "adfs_signing_cert"
  sp_entity_id  = "https://app.example.com/saml"
  sp_cert       = "useselfsigned"
}

# KCD outbound domain (server-side delegation)
resource "kemp_sso_domain" "kcd_out" {
  name            = "kcd.example.com"
  auth_type       = "KCD"
  server_side     = true
  kerberos_domain = "CORP.EXAMPLE.COM"
  kerberos_kdc    = "dc01.corp.example.com"
  kcd_username    = "svc-kemp@corp.example.com"
  kcd_password    = var.kcd_password
}

# Use domain on a virtual service (form-based SSO with LDAP)
resource "kemp_virtual_service" "app" {
  address  = "10.0.0.100"
  port     = "443"
  protocol = "tcp"
  type     = "http"

  ssl_acceleration = true
  cert_files       = ["app_cert"]

  esp_enabled         = true
  esp_sso_domain      = kemp_sso_domain.ldap.name
  esp_input_auth_mode = "form"
}
