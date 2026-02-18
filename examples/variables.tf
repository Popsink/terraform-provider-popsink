variable "kafka_password" {
  description = "Password for Kafka connectors"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "postgres_password" {
  description = "Password for the PostgreSQL source connector"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "ibmi_password" {
  description = "Password for the IBM i source connector"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "oracle_password" {
  description = "Password for the Oracle target connector"
  type        = string
  sensitive   = true
  default     = "default"
}
