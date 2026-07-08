resource "popsink_subscription" "orders_to_snowflake" {
  name                = "orders-to-snowflake"
  datamodel_id        = var.datamodel_id
  target_connector_id = popsink_connector.snowflake_target.id

  target_table_name = "ORDERS"
  backfill          = true
  desired_state     = "running"

  smt_config = jsonencode([
    { function = "rename", from = "cust_id", to = "customer_id" },
  ])
}
