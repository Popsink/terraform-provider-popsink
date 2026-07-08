resource "popsink_env" "example" {
  name = "production"

  retention_configuration = {
    bootstrap_server  = "kafka.example.com:9093"
    security_protocol = "SASL_SSL"
    sasl_mechanism    = "SCRAM-SHA-256"
    sasl_username     = var.broker_username
    sasl_password     = var.broker_password
  }
}
