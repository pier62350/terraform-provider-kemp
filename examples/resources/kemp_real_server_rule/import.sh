# Import format: <virtual_service_id>/<real_server_id>/<vs_port>/<vs_protocol>/<rs_address>/<rs_port>/<rule_name>
terraform import kemp_real_server_rule.example 12/5/80/tcp/192.168.1.10/8080/health-check
