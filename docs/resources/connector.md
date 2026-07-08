---
page_title: "popsink_connector Resource - Popsink"
subcategory: ""
description: |-
  Manages a Popsink connector (source or target).
---

# popsink_connector (Resource)

Manages a Popsink connector. A connector represents a data source or target integration (e.g., Kafka, Oracle, PostgreSQL, IBM i, Iceberg, Snowflake).

## Supported connector types

`connector_type` accepts any of the Popsink data-plane connector types. Every type
listed here can be created from Terraform.

**Sources**: `KAFKA_SOURCE`, `IBMI_SOURCE`, `ORACLE_SOURCE`, `POSTGRES_SOURCE`,
`MSSQL_SOURCE`, `MYSQL_SOURCE`, `DLT_SOURCE`, `ZENDESK_SOURCE`, `SHOPIFY_SOURCE`,
`PIPEDRIVE_SOURCE`, `HUBSPOT_SOURCE`, `GOOGLE_ADS_SOURCE`, `FACEBOOK_ADS_SOURCE`,
`SHAREPOINT_SOURCE`, `SALESFORCE_SOURCE`, `POPSINK_SOURCE`, `DATAGEN_SOURCE`,
`BIGQUERY_SOURCE`, `SNOWFLAKE_SOURCE`.

**Targets**: `ORACLE_TARGET`, `KAFKA_TARGET`, `ICEBERG_TARGET`,
`UNITY_CATALOG_TARGET`, `SNOWFLAKE_TARGET`, `POSTGRES_TARGET`,
`ELASTICSEARCH_TARGET`, `BIGQUERY_TARGET`, `WEBHOOK_TARGET`.

The set of fields inside `json_configuration` depends on the connector type. The
examples below cover a representative selection; SaaS sources (Zendesk, Shopify,
Pipedrive, Salesforce, Google Ads, Facebook Ads, SharePoint) follow the same
token/OAuth pattern as the HubSpot example.

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

### Oracle Source

```hcl
resource "popsink_connector" "oracle_source" {
  name           = "oracle-source"
  connector_type = "ORACLE_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host         = "oracle.example.com"
    port         = 1521
    service_name = "ORCLPDB1"
    user         = "logminer_user"
    password     = var.oracle_password
    whitelist    = "SALES.ORDERS,SALES.CUSTOMERS"
    init_load    = true
  })
}
```

### SQL Server Source

```hcl
resource "popsink_connector" "mssql_source" {
  name           = "mssql-source"
  connector_type = "MSSQL_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host                     = "mssql.example.com"
    port                     = 1433
    database                 = "production"
    user                     = "cdc_user"
    password                 = var.mssql_password
    encrypt                  = true
    trust_server_certificate = false
    whitelist                = "dbo.orders,dbo.customers"
    init_load                = true
  })
}
```

### MySQL Source

```hcl
resource "popsink_connector" "mysql_source" {
  name           = "mysql-source"
  connector_type = "MYSQL_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host      = "mysql.example.com"
    port      = 3306
    database  = "production"
    user      = "replication_user"
    password  = var.mysql_password
    whitelist = "production.orders,production.customers"
    init_load = true
  })
}
```

### HubSpot Source (SaaS / token-based)

```hcl
resource "popsink_connector" "hubspot_source" {
  name           = "hubspot-source"
  connector_type = "HUBSPOT_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    access_token = var.hubspot_access_token
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

### PostgreSQL Target

```hcl
resource "popsink_connector" "postgres_target" {
  name           = "postgres-target"
  connector_type = "POSTGRES_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host     = "postgres-dw.example.com"
    port     = "5432"
    database = "warehouse"
    user     = "loader"
    password = var.postgres_target_password
    schema   = "public"
  })
}
```

### Unity Catalog Target

```hcl
resource "popsink_connector" "unity_catalog_target" {
  name           = "unity-catalog-target"
  connector_type = "UNITY_CATALOG_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    auth_method               = "spn"
    tenant_id                 = var.azure_tenant_id
    client_id                 = var.azure_client_id
    client_secret             = var.azure_client_secret
    storage_account_name      = "mydatalake"
    container_name            = "delta"
    workspace_url             = "https://adb-1234567890.0.azuredatabricks.net"
    databricks_client_id      = var.databricks_client_id
    databricks_client_secret  = var.databricks_client_secret
    catalog_name              = "main"
    schema_name               = "cdc"
  })
}
```

### Elasticsearch Target

```hcl
resource "popsink_connector" "elasticsearch_target" {
  name           = "elasticsearch-target"
  connector_type = "ELASTICSEARCH_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    url         = "https://es.example.com:9243"
    auth_method = "api_key"
    api_key     = var.elasticsearch_api_key
    verify_ssl  = true
  })
}
```

### BigQuery Target

```hcl
resource "popsink_connector" "bigquery_target" {
  name           = "bigquery-target"
  connector_type = "BIGQUERY_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    service_account = var.gcp_service_account_json
    project         = "my-gcp-project"
    dataset         = "cdc"
  })
}
```

### Webhook Target

```hcl
resource "popsink_connector" "webhook_target" {
  name           = "webhook-target"
  connector_type = "WEBHOOK_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host                = "https://example.com/ingest"
    authentication_type = "BEARER"
    token               = var.webhook_token
    request_method      = "POST"
  })
}
```

## Argument Reference

- `name` (Required) - The name of the connector. Must only contain alphanumeric characters, hyphens, and underscores.
- `connector_type` (Required) - The type of connector. See [Supported connector types](#supported-connector-types) for the full list of accepted values.
- `json_configuration` (Required, Sensitive) - The connector configuration as a JSON string. Use `jsonencode()` for readability. Configuration fields depend on the connector type. This attribute is marked sensitive: its value is redacted from `terraform plan` output and logs. See [Handling secrets](#handling-secrets) below.
- `team_id` (Required) - The ID of the team that owns this connector.
- `desired_state` (Optional) - Desired lifecycle state of the connector worker: `running` or `stopped`. When set, the provider starts/stops the worker and waits for the status to converge. Omit to leave the worker in whatever state the API defaults to. Only applies to connector types with a controllable worker (e.g. Kafka sources configured for retention). See [Lifecycle control](#lifecycle-control-desired_state).
- `state_timeout` (Optional) - Maximum time to wait for the worker to reach `desired_state`, as a Go duration (e.g. `"5m"`, `"90s"`). Defaults to `"5m"`.

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

## Lifecycle control (`desired_state`)

Connector workers have their own lifecycle (start/stop). By default a connector
created through Terraform is left in whatever state the API assigns it. Set
`desired_state` to manage it declaratively:

```hcl
resource "popsink_connector" "kafka_source" {
  name           = "kafka-source"
  connector_type = "KAFKA_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    bootstrap_servers = "kafka.example.com:9092"
    topic             = "orders"
  })

  desired_state = "running"
  state_timeout = "10m"
}
```

Behavior:

- On create and update, if the worker is not already in `desired_state`, the
  provider calls the start/stop endpoint and **polls `GET /connectors/{id}`
  until the status converges** (`live` for `running`, `paused` for `stopped`),
  or `state_timeout` elapses (which fails the apply).
- It is a **no-op when the worker already matches** the desired state.
- Starting a worker that ends in `error` completes the apply with a warning
  (the start was issued but the worker did not become healthy) — inspect the
  worker logs.
- On read, `desired_state` is refreshed from the observed worker status, so a
  worker stopped out-of-band shows a diff on the next plan and is restarted.

The same start/stop + convergence pattern applies to
[subscriptions](subscription.md) via their own `desired_state`.

## Import

Connectors can be imported using their ID:

```bash
terraform import popsink_connector.example <connector-id>
```
