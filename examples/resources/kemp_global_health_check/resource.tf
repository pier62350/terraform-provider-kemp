resource "kemp_global_health_check" "this" {
  timeout     = 5
  retry_count = 3
  # retry_interval is auto-computed by the LoadMaster as retry_count * timeout + 1
}
