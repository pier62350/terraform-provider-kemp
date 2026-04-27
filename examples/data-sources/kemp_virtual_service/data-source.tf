data "kemp_virtual_service" "existing" {
  id = "12"
}

output "existing_address" {
  value = data.kemp_virtual_service.existing.address
}
