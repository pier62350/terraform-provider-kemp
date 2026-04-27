terraform {
  required_providers {
    kemp = {
      source  = "pier62350/kemp"
      version = "~> 0.1"
    }
  }
}

# All credentials are typically read from environment variables:
#   KEMP_HOST, KEMP_API_KEY (or KEMP_USERNAME/KEMP_PASSWORD)
provider "kemp" {
  host    = "https://10.0.0.5:9443"
  api_key = var.kemp_api_key
}

variable "kemp_api_key" {
  type      = string
  sensitive = true
}
