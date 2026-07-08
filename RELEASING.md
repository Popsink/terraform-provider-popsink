# Releasing

This provider is published to the [Terraform Registry](https://registry.terraform.io/)
under the `popsink` namespace. Releases are cut by pushing a `v*` tag, which
triggers the `release` GitHub Actions workflow (GoReleaser + GPG signing).

## One-time registry onboarding

These steps are performed once, by a `popsink` GitHub organization owner.

1. **Generate a GPG signing key** (RSA 4096, no expiry recommended for CI):

   ```bash
   gpg --full-generate-key
   gpg --armor --export-secret-keys <KEY_ID>   # private key (for the CI secret)
   gpg --armor --export <KEY_ID>                # public key (for the registry)
   ```

2. **Add the CI secrets** to the repository (Settings → Secrets and variables → Actions):
   - `GPG_PRIVATE_KEY` — the ASCII-armored **private** key.
   - `GPG_PASSPHRASE` — its passphrase.

   The `release` workflow imports these via `crazy-max/ghaction-import-gpg` and
   exposes `GPG_FINGERPRINT`, which GoReleaser uses to sign the checksums.

3. **Register the provider on the Terraform Registry:**
   - Sign in to <https://registry.terraform.io/> with the `popsink` GitHub org.
   - Publish a new provider and select this repository (the repo must be named
     `terraform-provider-popsink`).
   - Add the **public** GPG key under the org's signing keys.

## Cutting a release

1. Update `CHANGELOG.md`: move items from `## [Unreleased]` into a new
   `## [x.y.z]` section.
2. Ensure `go mod tidy` is clean and `make test` passes.
3. Tag and push:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

4. The `release` workflow runs GoReleaser and creates a **draft** GitHub release.
   Review it, then publish. The Terraform Registry picks up the published
   release automatically.

## What GoReleaser produces (registry-compatible artifacts)

`.goreleaser.yml` is configured to emit exactly what the registry expects:

- `terraform-provider-popsink_<version>_<os>_<arch>.zip` — one **zip** per
  platform (the registry requires zip for every OS, not tar.gz).
- `terraform-provider-popsink_<version>_SHA256SUMS` — checksums.
- `terraform-provider-popsink_<version>_SHA256SUMS.sig` — detached GPG
  signature of the checksums.
- `terraform-provider-popsink_<version>_manifest.json` — from
  `terraform-registry-manifest.json`, declaring `protocol_versions: ["6.0"]`
  (this provider is built on terraform-plugin-framework, protocol 6).

## Documentation

Docs live under `docs/` (`index.md`, `resources/*.md`, `data-sources/*.md`) with
the front matter the Terraform Registry renders. They are currently
**hand-maintained**: the CI `docs` job runs `go generate ./...`, but no
`//go:generate` directive is wired yet, so that step is a no-op placeholder.

Adopting [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) to
generate the docs from schema + `examples/` is an optional future enhancement.
To wire it, add a `tools.go` pinning the module and a `//go:generate
tfplugindocs generate` directive, then commit the generated output so the CI
diff check passes. Until then, update the `docs/` files by hand when a
resource's schema changes.
