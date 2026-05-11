# Terraform Provider for Kemp LoadMaster

A Terraform provider for managing [Kemp LoadMaster](https://www.progress.com/kemp) (Progress) resources via the LoadMaster API v2.

> Inspired by [`kreemer/terraform-provider-loadmaster`](https://github.com/kreemer/terraform-provider-loadmaster) and [`mathlu/terraform-provider-loadmaster`](https://github.com/mathlu/terraform-provider-loadmaster). The HTTP client and resource layer are written from scratch under MPL-2.0.

Published at [registry.terraform.io/providers/pier62350/kemp](https://registry.terraform.io/providers/pier62350/kemp).

## Requirements

- Terraform >= 1.0
- Go >= 1.24 (for building from source)
- A Kemp LoadMaster instance reachable over HTTPS with API access enabled

## Provider configuration

```hcl
terraform {
  required_providers {
    kemp = {
      source  = "pier62350/kemp"
      version = "~> 0.5"
    }
  }
}

provider "kemp" {
  host    = "https://10.0.0.5:9443"  # or use KEMP_HOST
  api_key = var.kemp_api_key         # or use KEMP_API_KEY

  # Alternative: basic auth
  # username = "bal"
  # password = var.kemp_password
}
```

Environment variables (preferred for credentials):

| Variable        | Purpose                                |
|-----------------|----------------------------------------|
| `KEMP_HOST`     | LoadMaster URL, e.g. `https://lm:9443` |
| `KEMP_API_KEY`  | API key for the LoadMaster user        |
| `KEMP_USERNAME` | Username (basic auth)                  |
| `KEMP_PASSWORD` | Password (basic auth)                  |

## Resources

### Virtual Services

| Resource / Data Source | Notes |
|---|---|
| `kemp_virtual_service` | VS with SSL/SNI, ESP, WAF intercept_mode |
| `kemp_sub_virtual_service` | Sub-VS (same surface as VS) |
| `kemp_real_server` | Backend pool member |
| `kemp_virtual_service_rule` | Attach a content rule to a VS |
| `kemp_sub_virtual_service_rule` | Attach a content rule to a SubVS |
| `kemp_virtual_service_waf_rule` | Attach a WAF rule (or rule set) to a VS |
| `kemp_real_server_rule` | Attach a content rule to a real server |
| `kemp_virtual_service_access_list_entry` | Per-VS block/allow list entry |

### Content Rules

| Resource | Notes |
|---|---|
| `kemp_match_content_rule` | Match rule (type 0) |
| `kemp_add_header_rule` | Add header (type 1) |
| `kemp_delete_header_rule` | Delete header (type 2) |
| `kemp_replace_header_rule` | Replace header (type 3) |
| `kemp_modify_url_rule` | Modify URL (type 4) |
| `kemp_replace_body_rule` | Replace body (type 5) |

### Certificates & Security

| Resource / Data Source | Notes |
|---|---|
| `kemp_certificate` | PEM/PFX upload |
| `kemp_certificates` (data) | List installed certificates |
| `kemp_acme_certificate` | Let's Encrypt / DigiCert issuance |
| `kemp_acme_certificate_renewal` | Force-renew an ACME cert |
| `kemp_acme_account` | Service-level ACME bootstrap (one-shot) |
| `kemp_cipher_set` | Custom TLS cipher sets |
| `kemp_sso_domain` | SSO domain (Kerberos, SAML, NTLM, …) |

### WAF

| Resource / Data Source | Notes |
|---|---|
| `kemp_owasp_custom_rule` | OWASP/ModSecurity custom rule file (admin-level) |
| `kemp_waf_custom_rule` | Legacy commercial WAF custom rule file |

### LoadMaster Configuration (`kemp_config_*`)

Global settings not tied to a specific VS or RS.

| Resource / Data Source | Notes |
|---|---|
| `kemp_config_global_health_check` | Global default health-check parameters |
| `kemp_config_route` | Static route |
| `kemp_config_hosts_entry` | `/etc/hosts`-style local resolution entry |
| `kemp_config_interface` (data) | Network interface details (read-only) |
| `kemp_config_ldap_endpoint` | LDAP server for authentication |
| `kemp_config_local_user` | Local WUI user account |
| `kemp_config_group` | User permission group |
| `kemp_config_user_certificate` | Per-user client certificate |
| `kemp_config_syslog` | Syslog destination per severity level |
| `kemp_config_ha` | HA pair setup (mode, partner IP, shared IP, secret) |
| `kemp_config_packet_routing_filter` | Packet routing filter |
| `kemp_config_intrusion_detection` | IPS/IDS paranoia level 0–4 |
| `kemp_config_access_list_entry` | Global block/allow list entry |
| `kemp_config_bandwidth_limit` | Global client bandwidth limit per IP |
| `kemp_config_cps_limit` | Global connections-per-second limit per IP |
| `kemp_config_rps_limit` | Global requests-per-second limit per IP |
| `kemp_config_connection_limit` | Global max concurrent connection limit per IP |
| `kemp_config_url_limit_rule` | URL-based rate limiting rule |
| `kemp_config_interface_vlan` | VLAN on a network interface |
| `kemp_config_interface_vxlan` | VXLAN tunnel |
| `kemp_config_interface_address` | Additional IP address on an interface |
| `kemp_config_interface_bond` | Bonded (LAG) interface |
| `kemp_config_cache_extension` | Per-extension cache bypass |
| `kemp_config_compression_extension` | Per-extension compression bypass |
| `kemp_config_network_telemetry` | Network telemetry enable/disable per interface |
| `kemp_config_geo` | Enable/disable GEO load balancing globally |
| `kemp_config_waf_update` | WAF auto-update and install schedule |
| `kemp_config_waf_logging` | WAF remote logging (URI, credentials, format) |
| `kemp_config_kubernetes` | Kubernetes ingress integration |

### GEO / Global Server Load Balancing (`kemp_gslb_*`)

| Resource / Data Source | Notes |
|---|---|
| `kemp_gslb_cluster` | GEO cluster (remote LoadMaster pool) |
| `kemp_gslb_location` | Custom geographic location for IP range selection |
| `kemp_gslb_fqdn` | FQDN with IP members and location weights |
| `kemp_gslb_params` | DNS zone / SOA parameters for GSLB (singleton) |
| `kemp_gslb_ip_range` | IP range to location mapping |
| `kemp_gslb_acl_entry` | GEO ACL custom entry (allow/block by CIDR) |

## Example

```hcl
resource "kemp_virtual_service" "web" {
  address  = "10.0.0.100"
  port     = "443"
  protocol = "tcp"
  type     = "http"
  nickname = "web"
  enabled  = true
}

resource "kemp_real_server" "web" {
  virtual_service_id = kemp_virtual_service.web.id
  address            = "10.0.0.10"
  port               = "8080"
  weight             = 1000
  enable             = true
}

resource "kemp_add_header_rule" "correlation_id" {
  name   = "add-correlation-id"
  header = "X-Correlation-ID"
  value  = "%[unique-id]"
}

resource "kemp_virtual_service_rule" "web" {
  virtual_service_id = kemp_virtual_service.web.id
  rule               = kemp_add_header_rule.correlation_id.name
  direction          = "request"
}
```

## Development

```bash
# Build
make build

# Unit tests
make test

# Acceptance tests (against a real LoadMaster — needs KEMP_HOST + KEMP_API_KEY)
make testacc

# Regenerate docs
make generate
```

A development override for `~/.terraformrc` lets Terraform pick up a local build:

```hcl
provider_installation {
  dev_overrides {
    "pier62350/kemp" = "/home/<user>/go/bin"
  }
  direct {}
}
```

## License

[MPL-2.0](./LICENSE)
