resource "popsink_team_member" "owner" {
  team_id = popsink_team.example.id
  user_id = var.owner_user_id
  role    = "owner"
}

resource "popsink_team_member" "member" {
  team_id = popsink_team.example.id
  user_id = var.member_user_id
  # role defaults to "member"
}
