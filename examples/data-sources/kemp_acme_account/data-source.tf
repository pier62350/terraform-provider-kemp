data "kemp_acme_account" "le" {
  acme_type = "1"
}

output "account_id" {
  value = data.kemp_acme_account.le.account_id
}
