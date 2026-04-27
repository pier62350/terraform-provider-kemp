# Terraform Provider for Kemp LoadMaster

A Terraform provider for managing [Kemp LoadMaster](https://www.progress.com/kemp) (Progress) resources via the LoadMaster API v2.

> Inspired by [`kreemer/terraform-provider-loadmaster`](https://github.com/kreemer/terraform-provider-loadmaster) and [`mathlu/terraform-provider-loadmaster`](https://github.com/mathlu/terraform-provider-loadmaster). The HTTP client and resource layer are written from scratch under MPL-2.0.

## Status

`v0.1.0` — early development. Resources currently implemented:

- `kemp_virtual_service` (resource + data source)
- `kemp_real_server` (resource + data source)
- `kemp_certificate` (resource)

Planned for `v0.2.0`: ESP (Edge Security Pack) configuration on virtual services, WAF / OWASP rules, header / URL rewrite rules, sub virtual services.

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
      version = "~> 0.1"
    }
  }
}

provider "kemp" {
  host    = "https://10.0.0.5:9443"  # or use KEMP_HOST
  api_key = var.kemp_api_key              # or use KEMP_API_KEY

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

## Example

```hcl
resource "kemp_virtual_service" "demo" {
  address  = "10.0.0.100"
  port     = "443"
  protocol = "tcp"
  type     = "http"
  nickname = "demo"
  enabled  = true
}

resource "kemp_real_server" "demo" {
  virtual_service_id = kemp_virtual_service.demo.id
  address            = "10.0.0.10"
  port               = "8080"
  weight             = 1000
  enable             = true
}
```

## Development

```bash
# Build
make build

# Lint
make lint

# Unit tests
make test

# Acceptance tests (against a real LoadMaster — needs KEMP_HOST + KEMP_API_KEY)
make testacc
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
