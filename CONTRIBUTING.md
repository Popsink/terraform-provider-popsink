# Contributing

## Development

```bash
make build    # lint + build
make test     # unit tests (excludes acceptance tests)
make testacc  # acceptance tests (requires a live data-plane, sets TF_ACC=1)
make docs     # regenerate docs
```

## Acceptance tests

Acceptance tests (`TestAcc*`) exercise the provider end-to-end against a **real
data-plane**: they create/update/import/destroy actual resources. They are
skipped unless `TF_ACC=1` (set by `make testacc`), so `make test` stays fast and
hermetic.

### Running locally

Point the provider at a data-plane you can safely write to and run `make testacc`:

```bash
export POPSINK_BASE_URL="https://data-plane.example.com/api"
export POPSINK_TOKEN="…"
export POPSINK_INSECURE="true"   # only for self-signed certs

make testacc                     # all acceptance tests
make testacc T=TestAccTeamResource   # a single test
```

Some tests need a pre-existing object the API cannot create from Terraform and
are **skipped** unless you provide it:

| Env var | Needed by |
|---------|-----------|
| `POPSINK_TEST_USER_ID` | `TestAccTeamMemberResource` (a user to add to a team) |
| `POPSINK_TEST_DATAMODEL_ID` + `POPSINK_TEST_TARGET_CONNECTOR_ID` | `TestAccSubscriptionResource` |

The self-contained tests (`popsink_team`, `popsink_connector`, `popsink_env`,
and the `popsink_team` data source, plus a 404-removal/"disappears" test) run
with just `POPSINK_BASE_URL` / `POPSINK_TOKEN`.

### CI

Acceptance tests are **not** part of the fast unit CI. Run them on a schedule and
before a release via a dedicated workflow with a staging data-plane's
credentials in repository secrets (`POPSINK_BASE_URL`, `POPSINK_TOKEN`). Suggested
`.github/workflows/acceptance.yml`:

```yaml
name: acceptance
on:
  schedule:
    - cron: "0 6 * * 1" # weekly, Monday 06:00 UTC
  workflow_dispatch: {}
permissions:
  contents: read
jobs:
  testacc:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: ".go-version"
          cache: true
      - run: make testacc
        env:
          TF_ACC: "1"
          POPSINK_BASE_URL: ${{ secrets.POPSINK_BASE_URL }}
          POPSINK_TOKEN: ${{ secrets.POPSINK_TOKEN }}
```

## Keeping `connector_type` in sync

The provider validates `connector_type` against an explicit list of accepted
values in `internal/provider/connector_resource.go` (`var connectorTypes`). This
list mirrors the data-plane `ConnectorType` enum
(`data-plane/back/domains/entities/connector_entity.py`).

**Decided drift-prevention strategy** (issue #24):

1. **Hand-maintained list, guarded by a unit test.** When the data-plane adds or
   removes a connector type, update `connectorTypes`. `TestConnectorTypesCoversDataPlaneEnum`
   guards the expected count so an accidental edit is visible in review.
2. **Opt-in drift check against a live data-plane.** Run

   ```bash
   POPSINK_BASE_URL=https://data-plane.example.com/api make check-connector-types
   ```

   It fetches the data-plane OpenAPI spec (`/openapi.json`), extracts the
   `ConnectorType` enum, and diffs it against `connectorTypes`, failing on any
   mismatch. This is intentionally **not** a blocking CI step yet: it needs a
   reachable data-plane and a stable spec.
3. **Long-term fix.** Generate `connectorTypes` from a *versioned, published*
   OpenAPI spec so the list can never drift. This is blocked on
   data-plane#2530 (publish a versioned OpenAPI spec). Once that ships, replace
   the hand-maintained list + opt-in check with generation in CI.

When you add a type, also add a per-type `json_configuration` example to
`docs/resources/connector.md` and `examples/connectors.tf` where the shape is
non-trivial (SaaS/token sources can share the existing HubSpot example).
