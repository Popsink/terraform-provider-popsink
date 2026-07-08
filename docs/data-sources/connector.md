---
page_title: "popsink_connector Data Source - Popsink"
subcategory: ""
description: |-
  Look up an existing Popsink connector by name.
---

# popsink_connector (Data Source)

Look up an existing Popsink connector by name.

## Example Usage

```hcl
data "popsink_connector" "snowflake" {
  name = "snowflake-target"
}

output "snowflake_connector_id" {
  value = data.popsink_connector.snowflake.id
}
```

## Argument Reference

- `name` (Required) - The name of the connector to look up.

## Attribute Reference

- `id` - The unique identifier of the connector.
- `connector_type` - The connector type.
- `team_id` - The team that owns the connector.

The lookup fails with a clear error if no connector matches the given name.
