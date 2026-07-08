# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `popsink_env` resource — manage environments (the foundational namespace).
- `popsink_subscription` resource — datamodel-to-target subscriptions with
  `desired_state` (running/paused) lifecycle control.
- `popsink_team_member` resource — declarative team membership.
- `desired_state` lifecycle control on `popsink_connector` (start/stop with
  convergence polling and a configurable `state_timeout`).
- Data sources `popsink_env`, `popsink_team`, `popsink_connector` and
  `popsink_pipeline` for looking up existing resources by name.
- `terraform-registry-manifest.json` and release wiring for Terraform Registry
  publishing.

### Changed

- `connector_type` now accepts all data-plane connector types (28) instead of 7.
- `json_configuration` on `popsink_connector` is marked sensitive so credentials
  are redacted from plan output and logs.

### Security

- Bumped `google.golang.org/grpc` to v1.79.3 and `golang.org/x/net` to v0.55.0
  to patch CVE-2026-33186 and CVE-2026-25680.
