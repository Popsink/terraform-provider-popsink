---
page_title: "popsink_env Data Source - Popsink"
subcategory: ""
description: |-
  Look up an existing Popsink environment by name.
---

# popsink_env (Data Source)

Look up an existing Popsink environment by name.

## Example Usage

```hcl
data "popsink_env" "production" {
  name = "production"
}

resource "popsink_team" "data_eng" {
  name        = "data-engineering"
  description = "Data engineering team"
  env_id      = data.popsink_env.production.id
}
```

## Argument Reference

- `name` (Required) - The name of the environment to look up.

## Attribute Reference

- `id` - The unique identifier of the environment.

The lookup fails with a clear error if no environment matches the given name.
