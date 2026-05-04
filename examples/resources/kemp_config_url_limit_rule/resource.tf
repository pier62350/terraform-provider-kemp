resource "kemp_config_url_limit_rule" "example" {
  name    = "ExampleRule"
  pattern = "=/test/a.html"
  limit   = 5
  match   = 0 # 0 = exact, 1 = prefix, 2 = regex
}
