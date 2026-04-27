# Local dev smoke test — meant to be run with a `dev_overrides` entry in
# ~/.terraformrc so terraform picks up the binary built by `go install`.
#
# Set the environment first:
#   set -a; source ~/.kemp_env; set +a
#
# Then:
#   terraform plan
#   terraform apply
#   terraform destroy
#
# Safety: this targets a brand-new virtual service. Never apply against
# addresses already in use on the target LoadMaster.

terraform {
  required_providers {
    kemp = {
      source = "pier62350/kemp"
    }
  }
}

provider "kemp" {
  # KEMP_HOST and KEMP_API_KEY are read from the environment.
}

resource "kemp_virtual_service" "smoke" {
  address  = "10.0.0.100"
  port     = "8080"
  protocol = "tcp"
  type     = "http"
  nickname = "tf-smoke"
  enabled  = true
}

resource "kemp_real_server" "smoke" {
  virtual_service_id = kemp_virtual_service.smoke.id
  address            = "10.0.1.10"
  port               = "8080"
  weight             = 1000
  enable             = true
}

output "vs_id" {
  value = kemp_virtual_service.smoke.id
}

output "rs_id" {
  value = kemp_real_server.smoke.id
}
