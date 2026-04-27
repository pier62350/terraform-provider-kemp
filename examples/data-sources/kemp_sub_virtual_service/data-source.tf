data "kemp_sub_virtual_service" "existing" {
  id = "42"
}

output "nickname" {
  value = data.kemp_sub_virtual_service.existing.nickname
}
