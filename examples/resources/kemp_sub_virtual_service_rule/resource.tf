# Content switching: route requests matching "/api/*" to the API SubVS.

resource "kemp_virtual_service" "web" {
  address  = "10.0.0.100"
  port     = "80"
  protocol = "tcp"
  type     = "http"
}

resource "kemp_sub_virtual_service" "api" {
  parent_id = kemp_virtual_service.web.id
  nickname  = "api"
}

resource "kemp_sub_virtual_service" "static" {
  parent_id = kemp_virtual_service.web.id
  nickname  = "static"
}

# Match content rule that matches requests with path starting with /api/
resource "kemp_match_content_rule" "api_path" {
  name        = "match-api-path"
  match_type  = "prefix"
  header      = "url"
  replacement = "/api/"
}

# Attach the rule to the API SubVS so the parent VS routes /api/* requests to it.
resource "kemp_sub_virtual_service_rule" "api_routing" {
  parent_virtual_service_id = kemp_virtual_service.web.id
  sub_virtual_service_id    = kemp_sub_virtual_service.api.id
  rule                      = kemp_match_content_rule.api_path.name
}
