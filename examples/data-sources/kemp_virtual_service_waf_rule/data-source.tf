data "kemp_virtual_service_waf_rule" "existing" {
  virtual_service_address  = "10.0.0.100"
  virtual_service_port     = "443"
  virtual_service_protocol = "tcp"
  rule                     = "G/ip_reputation"
}
