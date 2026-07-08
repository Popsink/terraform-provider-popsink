# --- Sources ---

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

  # Declaratively run the worker and wait for it to converge.
  desired_state = "running"
  state_timeout = "10m"
}

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

resource "popsink_connector" "oracle_source" {
  name           = "oracle-source"
  connector_type = "ORACLE_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    host         = "oracle.example.com"
    port         = 1521
    service_name = "ORCLPDB1"
    user         = "logminer_user"
    password     = var.oracle_source_password
    whitelist    = "SALES.ORDERS,SALES.CUSTOMERS"
    init_load    = true
  })
}

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

resource "popsink_connector" "hubspot_source" {
  name           = "hubspot-source"
  connector_type = "HUBSPOT_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    access_token = var.hubspot_access_token
  })
}

# --- Targets ---

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

resource "popsink_connector" "iceberg_target" {
  name           = "iceberg-target"
  connector_type = "ICEBERG_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    url = "s3://data-lake/warehouse"
  })
}

resource "popsink_connector" "snowflake_target" {
  name           = "snowflake-target"
  connector_type = "SNOWFLAKE_TARGET"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({})
}

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
