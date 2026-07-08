---
page_title: "popsink_team Data Source - Popsink"
subcategory: ""
description: |-
  Look up an existing Popsink team by name.
---

# popsink_team (Data Source)

Look up an existing Popsink team by name — useful for referencing teams created
outside Terraform (UI or API) without importing them.

## Example Usage

```hcl
data "popsink_team" "data_eng" {
  name = "data-engineering"
}

resource "popsink_connector" "pg" {
  name           = "pg-source"
  connector_type = "POSTGRES_SOURCE"
  team_id        = data.popsink_team.data_eng.id
  # ...
}
```

## Argument Reference

- `name` (Required) - The name of the team to look up.

## Attribute Reference

- `id` - The unique identifier of the team.
- `description` - The team description.
- `env_id` - The environment the team belongs to.

The lookup fails with a clear error if no team matches the given name.
