resource "kemp_config_syslog" "notice" {
  level = "notice"
  hosts = ["10.0.0.5:514", "10.0.0.6:514"]
}

resource "kemp_config_syslog" "error" {
  level = "error"
  hosts = ["10.0.0.5:514"]
}
