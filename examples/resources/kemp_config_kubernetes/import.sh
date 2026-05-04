# Singleton — import with the fixed ID "loadmaster".
# Note: kubeconfig_base64 is write-only and will not be populated; run
# terraform apply after import to reconcile.
terraform import kemp_config_kubernetes.main loadmaster
