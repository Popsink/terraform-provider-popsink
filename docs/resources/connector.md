---
page_title: "popsink_connector Resource - Popsink"
subcategory: ""
description: |-
  Manages a Popsink connector (source or target).
---

# popsink_connector (Resource)

Manages a Popsink connector. A connector represents a data source or target integration (e.g., Kafka, Oracle, PostgreSQL, IBM i, Iceberg, Snowflake).

## Example Usage

### Kafka Source

```hcl
resource "popsink_connector" "kafka_source" {
  name           = "kafka-source"
  connector_type = "KAFKA_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    bootstrap_servers = "kafka.example.com:9092"
    security_protocol = "SASL_SSL"
    sasl_mechanism    = "SCRAM-SHA-256"
    sasl_username     = "admin"
    sasl_password     = var.kafka_password
    topic             = "orders"
  })
}
```

### PostgreSQL Source

```hcl
resource "popsink_connector" "postgres_source" {
  name           = "postgres-source"
  connector_type = "POSTGRES_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host      = "postgres.example.com"
    port      = "5432"
    database  = "production"
    user      = "replication_user"
    password  = var.postgres_password
    whitelist = "public.orders,public.customers"
    init_load = true
  })
}
```

### IBM i Source

```hcl
resource "popsink_connector" "ibmi_source" {
  name           = "ibmi-source"
  connector_type = "IBMI_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host      = "ibmi.example.com"
    port      = "446"
    user      = "QUSER"
    password  = var.ibmi_password
    schema    = "MYLIB"
    whitelist = "ORDERS,CUSTOMERS"
    init_load = true
  })
}
```

### Oracle Target

```hcl
resource "popsink_connector" "oracle_target" {
  name           = "oracle-target"
  connector_type = "ORACLE_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host     = "oracle.example.com"
    port     = 1521
    database = "ORCL"
    username = "admin"
    password = var.oracle_password
    schema   = "DATA"
    table    = "ORDERS"
  })
}
```

### Kafka Target

```hcl
resource "popsink_connector" "kafka_target" {
  name           = "kafka-target"
  connector_type = "KAFKA_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    bootstrap_servers = "kafka.example.com:9092"
    security_protocol = "SASL_SSL"
    sasl_mechanism    = "SCRAM-SHA-256"
    sasl_username     = "admin"
    sasl_password     = var.kafka_password
    topic             = "orders-replicated"
  })
}
```

### Iceberg Target

```hcl
resource "popsink_connector" "iceberg_target" {
  name           = "iceberg-target"
  connector_type = "ICEBERG_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    url = "s3://data-lake/warehouse"
  })
}
```

### Snowflake Target

```hcl
resource "popsink_connector" "snowflake_target" {
  name           = "snowflake-target"
  connector_type = "SNOWFLAKE_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({})
}
```

## Argument Reference

- `name` (Required) - The name of the connector. Must only contain alphanumeric characters, hyphens, and underscores.
- `connector_type` (Required) - The type of connector. Valid values: `KAFKA_SOURCE`, `IBMI_SOURCE`, `POSTGRES_SOURCE`, `ORACLE_TARGET`, `KAFKA_TARGET`, `ICEBERG_TARGET`, `SNOWFLAKE_TARGET`.
- `json_configuration` (Required) - The connector configuration as a JSON string. Use `jsonencode()` for readability. Configuration fields depend on the connector type.
- `team_id` (Required) - The ID of the team that owns this connector.

## Attribute Reference

- `id` - The unique identifier of the connector.
- `items_count` - The number of items associated with this connector.
- `status` - The current status of the connector.

## Import

Connectors can be imported using their ID:

```bash
terraform import popsink_connector.example <connector-id>
```
