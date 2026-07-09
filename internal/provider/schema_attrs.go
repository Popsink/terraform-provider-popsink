package provider

// Shared schema attribute names used across more than one resource or data
// source. Used for schema map keys, path lookups, and structured-log fields.
//
// Note: struct `tfsdk:"…"` tags must stay string literals — Go struct tags
// cannot reference constants — so these constants intentionally cover only the
// non-tag occurrences.
const (
	attrTeamID       = "team_id"
	attrDesiredState = "desired_state"
)
