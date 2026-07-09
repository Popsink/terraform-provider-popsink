---
page_title: "popsink_datamodel Resource - Popsink"
subcategory: ""
description: |-
  Adopts an existing Popsink datamodel and manages its lifecycle.
---

# popsink_datamodel (Resource)

Adopts an existing Popsink datamodel and manages its lifecycle.

Datamodels have **no create endpoint** — they are derived from pipeline/connector
configuration. This resource references an existing datamodel by ID (similar to
`aws_default_vpc`): it does **not** create or delete the underlying datamodel,
only manages its desired lifecycle state. Destroying the resource removes it from
Terraform state and leaves the datamodel untouched.

## Example Usage

```hcl
# The datamodel is created by the data-plane once its source connector exists.
data "popsink_connector" "orders_source" {
  name = "orders-source"
}

resource "popsink_datamodel" "orders" {
  datamodel_id  = "b1e6f0c2-1111-2222-3333-444455556666"
  desired_state = "running"
}
```

## Argument Reference

- `datamodel_id` (Required) - The ID of the existing datamodel to adopt. Changing this adopts a different datamodel (forces replacement).
- `desired_state` (Optional) - Lifecycle state: `running` (enabled) or `stopped` (disabled). Managed via the datamodel start/stop endpoints. Defaults to the server state.

## Attribute Reference

- `id` - The datamodel identifier (mirrors `datamodel_id`).
- `name` - The datamodel name.
- `state` - The datamodel's current worker state.

## Behavior notes

- There is no create endpoint: the datamodel must already exist (create its
  source connector/pipeline first). Adopting a non-existent ID fails the apply
  with a clear error.
- `desired_state` is reconciled by calling the start/stop endpoints and is
  derived from the datamodel's `enabled` flag on read, so out-of-band changes
  are detected on the next plan.
- Destroying the resource is a no-op against the API — the datamodel is only
  released from Terraform state.

## Import

```bash
terraform import popsink_datamodel.orders <datamodel-id>
```
