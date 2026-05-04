resource "kemp_config_kubernetes" "main" {
  kubeconfig_base64 = filebase64("~/.kube/config")
  mode              = "active"
  namespace         = "default"
  watch_timeout     = 30
}
