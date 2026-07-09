# Datamodels are derived by the data-plane from a source connector; adopt one by
# ID to manage its lifecycle. It is never created or deleted by Terraform.
resource "popsink_datamodel" "orders" {
  datamodel_id  = var.datamodel_id
  desired_state = "running"
}
