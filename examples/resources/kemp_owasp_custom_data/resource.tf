resource "kemp_owasp_custom_data" "blocked_user_agents" {
  filename = "blocked_user_agents.data"
  data     = base64encode(file("${path.module}/blocked_user_agents.data"))
}
