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
- `json_configuration` (Required, Sensitive) - The connector configuration as a JSON string. Use `jsonencode()` for readability. Configuration fields depend on the connector type. This attribute is marked sensitive: its value is redacted from `terraform plan` output and logs. See [Handling secrets](#handling-secrets) below.
- `team_id` (Required) - The ID of the team that owns this connector.

## Attribute Reference

- `id` - The unique identifier of the connector.
- `items_count` - The number of items associated with this connector.
- `status` - The current status of the connector.

## Handling secrets

`json_configuration` typically carries credentials — database passwords, SASL
secrets, API keys. This section describes how those credentials are protected and
the recommended way to supply them.

### Sensitive marking

The attribute is marked `Sensitive`, so its value is **redacted from
`terraform plan` output and CI logs**. Terraform still records the value in state,
so protect your state backend accordingly (encryption at rest, restricted access).

### Recommended: pass credentials through sensitive variables

Never hard-code secrets in `.tf` files. Declare them as sensitive input variables
and reference them from `jsonencode(...)`:

```hcl
variable "postgres_password" {
  type      = string
  sensitive = true
}

resource "popsink_connector" "postgres_source" {
  name           = "postgres-source"
  connector_type = "POSTGRES_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host     = "postgres.example.com"
    port     = "5432"
    database = "production"
    user     = "replication_user"
    password = var.postgres_password
  })
}
```

Supply the value via `TF_VAR_postgres_password`, a `*.auto.tfvars` file kept out of
version control, or your CI secret store.

### Preferred: `valueFrom` secret references (credentials never enter state)

The Popsink data-plane supports referencing credentials from a Kubernetes Secret
instead of embedding the literal value. When a configuration field uses a
`valueFrom` reference, the credential value never reaches Terraform — it is
resolved by the data-plane at worker launch — so it is never written to plan
output *or* state:

```hcl
json_configuration = jsonencode({
  host     = "postgres.example.com"
  user     = "replication_user"
  password = { valueFrom = { secretKeyRef = { name = "pg-creds", key = "password" } } }
})
```

Use `valueFrom` wherever your deployment provisions secrets out of band; fall back
to inline sensitive variables only when no suitable Secret exists.

### Read-back behavior and drift

`GET /connectors/{id}` **redacts inline credentials** in its response: each
sensitive value is replaced by the fixed sentinel `"<redacted>"`, and the API
exposes a stable `config_hash` computed over the original (unredacted)
configuration for drift detection. Because the server never returns the real
credential, the provider keeps the `json_configuration` value from your
configuration in state rather than overwriting it with the redacted read-back —
this avoids a perpetual credential diff. A consequence is that **drift in
credential fields performed outside Terraform is not detected by value
comparison**; `config_hash`-based drift detection is tracked as a follow-up.

### Write-only attributes — evaluated, not adopted (yet)

Terraform 1.11+ / plugin-framework ≥1.15 support
[write-only arguments](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments),
which are never persisted to state. We evaluated exposing a
`json_configuration_wo` variant and decided **not** to adopt it in this release:

- The recommended `valueFrom` pattern already keeps credentials out of state
  entirely, which is a stronger guarantee than write-only (the value never even
  reaches the provider).
- A write-only config would force pairing with a server-provided drift signal
  (`config_hash`) to remain usable, which is a larger change tracked separately.

This decision will be revisited alongside `config_hash`-based drift detection.

## Import

Connectors can be imported using their ID:

```bash
terraform import popsink_connector.example <connector-id>
```
