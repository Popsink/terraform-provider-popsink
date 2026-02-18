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
