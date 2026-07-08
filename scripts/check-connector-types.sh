#!/usr/bin/env bash
#
# check-connector-types.sh — detect drift between the provider's connector_type
# list and the live data-plane ConnectorType enum.
#
# The provider maintains the accepted connector types by hand in
# internal/provider/connector_resource.go (var connectorTypes). This script
# compares that list against a running data-plane's OpenAPI spec so drift is
# caught explicitly rather than silently.
#
# Usage:
#   POPSINK_BASE_URL=https://data-plane.example.com/api ./scripts/check-connector-types.sh
#
# Optional:
#   POPSINK_TOKEN     bearer token, if the spec endpoint requires auth
#   POPSINK_INSECURE  set to "true" to skip TLS verification (self-signed certs)
#
# Exits non-zero when the two sets differ (missing or extra types), printing the
# diff. This is opt-in and not wired into CI blocking until the data-plane
# publishes a versioned, stable OpenAPI spec (data-plane#2530).

set -euo pipefail

: "${POPSINK_BASE_URL:?POPSINK_BASE_URL must be set (e.g. https://data-plane.example.com/api)}"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_file="$repo_root/internal/provider/connector_resource.go"

curl_args=(--fail --silent --show-error)
[ "${POPSINK_INSECURE:-}" = "true" ] && curl_args+=(--insecure)
[ -n "${POPSINK_TOKEN:-}" ] && curl_args+=(--header "Authorization: Bearer ${POPSINK_TOKEN}")

spec="$(curl "${curl_args[@]}" "${POPSINK_BASE_URL%/}/openapi.json")"

# ConnectorType enum values from the OpenAPI spec.
remote="$(printf '%s' "$spec" | python3 -c '
import json, sys
spec = json.load(sys.stdin)
schemas = spec.get("components", {}).get("schemas", {})
ct = schemas.get("ConnectorType")
if not ct or "enum" not in ct:
    sys.exit("ConnectorType enum not found in OpenAPI spec")
print("\n".join(sorted(ct["enum"])))
')"

# connectorTypes list from the Go source (quoted string literals between the
# "var connectorTypes" declaration and its closing brace).
local_types="$(awk '
  /var connectorTypes = \[\]string\{/ {inside=1; next}
  inside && /^\}/ {inside=0}
  inside {
    if (match($0, /"[A-Z_]+"/)) {
      s = substr($0, RSTART+1, RLENGTH-2); print s
    }
  }
' "$go_file" | sort)"

if [ "$remote" = "$local_types" ]; then
  echo "connector_type list is in sync with the data-plane ($(printf '%s' "$remote" | grep -c . ) types)."
  exit 0
fi

echo "connector_type drift detected between provider and data-plane:" >&2
echo "  '<' = only in provider   '>' = only in data-plane" >&2
diff <(printf '%s\n' "$local_types") <(printf '%s\n' "$remote") >&2 || true
exit 1
