---
page_title: "popsink_subscription Resource - Popsink"
subcategory: ""
description: |-
  Manages a Popsink subscription.
---

# popsink_subscription (Resource)

Manages a Popsink subscription. A subscription maps a datamodel to a target
[connector](connector.md) (optionally through SMT transformations) and is
managed independently of the pipeline composite.

## Example Usage

```hcl
resource "popsink_subscription" "orders_to_snowflake" {
  name                = "orders-to-snowflake"
  datamodel_id        = "b1e6f0c2-1111-2222-3333-444455556666"
  target_connector_id = popsink_connector.snowflake_target.id

  target_table_name = "ORDERS"
  backfill          = true
  desired_state     = "running"
}
```

### With an SMT transform chain

```hcl
resource "popsink_subscription" "orders_transformed" {
  name                = "orders-transformed"
  datamodel_id        = "b1e6f0c2-1111-2222-3333-444455556666"
  target_connector_id = popsink_connector.snowflake_target.id

  smt_config = jsonencode([
    { function = "rename", from = "cust_id", to = "customer_id" },
  ])

  error_table_enabled = true
  error_table_name    = "orders_errors"
  desired_state       = "paused"
}
```

## Argument Reference

- `name` (Required) - The name of the subscription.
- `datamodel_id` (Required) - The datamodel this subscription reads from. Changing this forces a new subscription.
- `target_connector_id` (Required) - The target connector this subscription delivers to. Changing this forces a new subscription.
- `smt_config` (Optional) - SMT (single-message transform) chain configuration as a JSON array string (use `jsonencode()`). Treated opaquely in v1; drift is tracked via `config_hash`.
- `consumer_id` (Optional) - Consumer ID for the subscription.
- `error_table_enabled` (Optional) - Whether an error table is enabled.
- `error_table_name` (Optional) - Name of the error table (when enabled).
- `error_table_target_id` (Optional) - Target connector ID that receives error-table rows.
- `target_table_name` (Optional) - Target table name. Defaults to `root`.
- `backfill` (Optional) - Whether the subscription backfills existing data.
- `desired_state` (Optional) - Lifecycle state: `running` (enabled) or `paused` (disabled). Defaults to the server state after creation. Managed via the subscription start/pause endpoints.

## Attribute Reference

- `id` - The unique identifier of the subscription.
- `status` - The current runtime status of the subscription (e.g. `live`, `paused`, `building`, `error`).
- `config_hash` - Server-computed stable hash of the subscription configuration, used for drift detection.

## Note on SMT configuration (v1)

`smt_config` is intentionally an opaque JSON array in v1 (typed attributes are a
future enhancement). The data-plane accepts the chain under `smt_config` on
create and `mapper_config` on update; the provider maps its single `smt_config`
attribute onto the appropriate field and keeps the configured value in state,
relying on `config_hash` for drift detection rather than comparing raw JSON.

## Import

Subscriptions can be imported using their ID:

```bash
terraform import popsink_subscription.example <subscription-uuid>
```
