# Contributing

## Development

```bash
make build    # lint + build
make test     # unit tests (excludes acceptance tests)
make testacc  # acceptance tests (requires a live data-plane, sets TF_ACC=1)
make docs     # regenerate docs
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
