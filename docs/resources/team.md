---
page_title: "popsink_team Resource - Popsink"
subcategory: ""
description: |-
  Manages a Popsink team.
---

# popsink_team (Resource)

Manages a Popsink team. Teams are used to organize connectors within an environment.

## Example Usage

```hcl
resource "popsink_team" "example" {
  name        = "data-engineering"
  description = "Data engineering team"
}
```

## Argument Reference

- `name` (Required) - The name of the team.
- `description` (Required) - Short description of the team.
- `env_id` (Optional) - The UUID of the environment the team is associated with.

## Attribute Reference

- `id` - The unique identifier of the team.

## Import

Teams can be imported using their ID:

```bash
terraform import popsink_team.example <team-id>
```
