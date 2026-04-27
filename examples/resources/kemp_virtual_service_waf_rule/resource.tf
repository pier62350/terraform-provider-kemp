resource "kemp_virtual_service" "web" {
  address            = "10.0.0.100"
  port               = "443"
  protocol           = "tcp"
  type               = "http"
  ssl_acceleration   = true
  waf_intercept_mode = "2" # OWASP WAF
}

# Attach a WAF rule set to the VS. The vsaddwafrule command keys on the
# VS triplet, not the Index.
resource "kemp_virtual_service_waf_rule" "ip_reputation" {
  virtual_service_address  = kemp_virtual_service.web.address
  virtual_service_port     = kemp_virtual_service.web.port
  virtual_service_protocol = kemp_virtual_service.web.protocol
  rule                     = "G/ip_reputation"
}

# Attach multiple rules in one go (space-percent-separated):
# resource "kemp_virtual_service_waf_rule" "core_set" {
#   virtual_service_address  = kemp_virtual_service.web.address
#   virtual_service_port     = kemp_virtual_service.web.port
#   virtual_service_protocol = kemp_virtual_service.web.protocol
#   rule                     = "G/malware_detection%20G/known_vulns"
#   disabled_rules           = "2200005,2200006"
# }
