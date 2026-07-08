---
page_title: "popsink_team_member Resource - Popsink"
subcategory: ""
description: |-
  Manages membership of a user in a Popsink team.
---

# popsink_team_member (Resource)

Manages membership of a user in a Popsink [team](team.md), so team access
control is reproducible across environments instead of only manageable in the
UI/API.

There is **no update endpoint** for a membership: changing the team, user, or
role replaces the resource (the old membership is removed and a new one added).

## Example Usage

```hcl
resource "popsink_team" "data_eng" {
  name        = "data-engineering"
  description = "Data engineering team"
}

resource "popsink_team_member" "alice" {
  team_id = popsink_team.data_eng.id
  user_id = "d290f1ee-6c54-4b01-90e6-d701748f0851"
  role    = "owner"
}

resource "popsink_team_member" "bob" {
  team_id = popsink_team.data_eng.id
  user_id = "a7d1f2fc-6e92-4dcd-b1f6-4200e4e9f1f3"
  # role defaults to "member"
}
```

## Argument Reference

- `team_id` (Required) - The ID of the team. Changing this forces a new membership.
- `user_id` (Required) - The ID of the user to add. Changing this forces a new membership.
- `role` (Optional) - `member` or `owner` (owners have admin privileges). Defaults to `member`. Changing this forces a new membership.

## Attribute Reference

- `id` - The membership entry identifier (distinct from `user_id`), used for removal and import.

## Behavior notes

- Membership is created through the team's bulk-add endpoint with a single user.
  A user who is already a member is skipped server-side; the resource still
  adopts the existing membership entry.
- Removing the sole remaining admin of a team while other members remain is
  rejected by the API (409) — Terraform surfaces the error. Removing the last
  member cascades to deleting the team itself (data-plane behavior).

## Import

Memberships can be imported using a composite `team_id/member_id` ID (where
`member_id` is the membership entry `id`, not the user ID):

```bash
terraform import popsink_team_member.alice <team-id>/<member-id>
```

## Environment members

An equivalent `popsink_env_member` resource is **not** currently offered: the
data-plane env-member API only exposes list/delete (membership is granted via
access requests), so environment membership cannot be created declaratively yet.
This is tracked upstream (data-plane#2526).
