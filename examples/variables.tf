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

variable "oracle_source_password" {
  description = "Password for the Oracle source connector"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "mssql_password" {
  description = "Password for the SQL Server source connector"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "hubspot_access_token" {
  description = "Private app access token for the HubSpot source connector"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "postgres_target_password" {
  description = "Password for the PostgreSQL target connector"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "webhook_token" {
  description = "Authentication token for the webhook target connector"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "broker_username" {
  description = "SASL username for the environment's Kafka broker"
  type        = string
  sensitive   = true
  default     = "default"
}

variable "broker_password" {
  description = "SASL password for the environment's Kafka broker"
  type        = string
  sensitive   = true
  default     = "default"
}
