# Reference existing Popsink resources created outside Terraform (UI/API)
# without importing them.

data "popsink_env" "production" {
  name = "production"
}

data "popsink_team" "data_eng" {
  name = "data-engineering"
}

data "popsink_connector" "snowflake" {
  name = "snowflake-target"
}

data "popsink_pipeline" "orders" {
  name = "orders-pipeline"
}

# Example: attach a new team to an existing environment.
resource "popsink_team" "analytics" {
  name        = "analytics"
  description = "Analytics team"
  env_id      = data.popsink_env.production.id
}
