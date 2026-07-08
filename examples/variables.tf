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

variable "datamodel_id" {
  description = "ID of the datamodel the example subscription reads from"
  type        = string
  default     = "00000000-0000-0000-0000-000000000000"
}

variable "owner_user_id" {
  description = "User ID to add to the example team as an owner"
  type        = string
  default     = "d290f1ee-6c54-4b01-90e6-d701748f0851"
}

variable "member_user_id" {
  description = "User ID to add to the example team as a member"
  type        = string
  default     = "a7d1f2fc-6e92-4dcd-b1f6-4200e4e9f1f3"
}
