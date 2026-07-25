# Releasing

## Prerequisites

- The public repositories `anirudh-777/homebrew-tap` and
  `anirudh-777/scoop-bucket` must exist.
- Add repository secrets `HOMEBREW_TAP_TOKEN` and `SCOOP_BUCKET_TOKEN` before
  publishing stable releases. Prerelease tags skip both package repositories.
- Add repository secret `NPM_TOKEN` to publish npm packages. Without it, the
  release still publishes GitHub artifacts and reports that npm was skipped.
- Configure Apple code signing and notarization before advertising the
  Homebrew cask as a stable macOS installation path. Sigstore verification
  does not replace Gatekeeper notarization.
- Confirm that `pb-agent` and all six `pb-agent-<platform>-<arch>` package
  names are controlled by the maintainer on npm.

## Release

1. Confirm `main` is green in CI and PocketBase compatibility checks.
2. Run `go test -race ./...`, `go vet ./...`, and `govulncheck ./...`.
3. Create and push an annotated SemVer tag such as `v0.1.0`.
4. Verify the GitHub release contains six archives, checksums, SBOMs, a
   Sigstore bundle, and provenance.
5. Install through npm, Homebrew, Scoop, and a release archive in clean
   environments.

Prereleases can be published to GitHub without package-manager credentials.
Before creating a stable release tag, configure every prerequisite secret and
verify all advertised installation paths.
