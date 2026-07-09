---
page_title: "popsink_pipeline Data Source - Popsink"
subcategory: ""
description: |-
  Look up an existing Popsink pipeline by name.
---

# popsink_pipeline (Data Source)

Look up an existing Popsink pipeline by name. The lookup is scoped to the API
token's active environment.

## Example Usage

```hcl
data "popsink_pipeline" "orders" {
  name = "orders-pipeline"
}

output "orders_pipeline_id" {
  value = data.popsink_pipeline.orders.id
}
```

## Argument Reference

- `name` (Required) - The name of the pipeline to look up.

## Attribute Reference

- `id` - The unique identifier of the pipeline.

The lookup fails with a clear error if no pipeline matches the given name.
