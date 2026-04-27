data "kemp_real_server" "existing" {
  virtual_service_id = "12"
  id                 = 26
}

output "existing_rs_address" {
  value = data.kemp_real_server.existing.address
}
