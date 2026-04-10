# Release Process

This document describes the minimal release process for BlockGo.

The goal is to keep releases:

- simple
- repeatable
- easy to review
- easy to reproduce
- aligned with the project's educational scope

## Release Scope

BlockGo is currently release-ready for educational `v0.x` releases.

A release should represent a repository state that has:

- passing formatting, vet, test, and build checks
- accurate documentation
- aligned config examples
- working local Docker demo flow
- clear version metadata in built binaries

This project is not claiming production-grade blockchain readiness. Releases should reflect the current educational scope honestly.

## Versioning

BlockGo should use semantic versioning tags in the form:

- `v0.1.0`
- `v0.1.1`
- `v0.2.0`

Recommended interpretation while the project is still educational and pre-1.0:

- patch release: docs fixes, small bug fixes, non-breaking polish
- minor release: meaningful new milestone-level functionality or notable UX improvements
- major release: reserved for a future stable `v1.0.0` boundary

## Release Artifacts

At minimum, a release should include:

- a Git tag
- GitHub release notes
- source code snapshot from the tagged commit

Optional future artifacts:

- prebuilt binaries
- checksums
- container image publishing

## Pre-Release Checklist

Before cutting a release, verify all of the following.

### Repository Quality

- [ ] `make fmt-check`
- [ ] `make vet`
- [ ] `make test`
- [ ] `make build`
- [ ] `make ci`

### Local Demo Validation

- [ ] `go run ./cmd/blockgo gen-localnet -out ./configs/local`
- [ ] `docker compose up --build` starts all nodes successfully
- [ ] `curl http://localhost:8001/healthz` returns success
- [ ] `curl http://localhost:8002/healthz` returns success
- [ ] `curl http://localhost:8003/healthz` returns success

### Documentation and Examples

- [ ] `README.md` matches the current codebase
- [ ] `docs/` is up to date
- [ ] `configs/*.example.json` matches current config behavior
- [ ] CLI examples still work
- [ ] Docker and Compose instructions still work

### Release Review

- [ ] version tag selected
- [ ] release notes drafted
- [ ] notable limitations and non-goals are still documented honestly
- [ ] no generated local secrets or data directories are staged for commit

## Recommended Release Workflow

### 1. Start from a clean main branch

Make sure your local branch is up to date and clean.

Example:

```bash
git checkout main
git pull --ff-only
git status
```

You should have no unintended local changes before preparing a release.

### 2. Run the full validation flow

Run the standard repository checks:

```bash
make fmt-check
make vet
make test
make build
make ci
```

Then validate the local multi-node demo:

```bash
go run ./cmd/blockgo gen-localnet -out ./configs/local
docker compose up --build
curl http://localhost:8001/healthz
curl http://localhost:8002/healthz
curl http://localhost:8003/healthz
```

After validation, stop the demo and clean up any generated local state if needed.

Example:

```bash
docker compose down -v
rm -rf ./configs/local ./data
```

Be careful not to remove anything you still need locally.

### 3. Review the release contents

Review what changed since the previous release.

Example:

```bash
git log --oneline <previous-tag>..HEAD
git diff --stat <previous-tag>..HEAD
```

Focus on:

- user-facing behavior changes
- docs changes
- config changes
- CI or Docker changes
- known limitations worth calling out

### 4. Create the release tag

Create an annotated tag for the release.

Example:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
```

Then push the tag:

```bash
git push origin v0.1.0
```

## Version Metadata in Binaries

The Makefile already injects version metadata into binaries through linker flags.

To build a release binary locally with a specific version string:

```bash
make build VERSION=v0.1.0
```

You can verify the embedded version with:

```bash
./bin/blockgo version
./bin/blockgo-node -version
```

Expected output shape:

```bash
blockgo v0.1.0 (commit=<commit> date=<timestamp>)
```

## GitHub Release Notes Template

Use concise release notes that reflect the actual scope.

Suggested structure:

```bash
## Summary

Short description of the release.

## Highlights

- item
- item
- item

## Validation

- `make ci`
- local Docker demo verified
- docs and config examples reviewed

## Known Limitations

- fixed validator set
- static peers
- educational scope, not production hardened
```

## What to Include in Release Notes

Good release notes usually mention:

- the release version
- the main improvements
- any fixes to docs, CI, Docker, or config
- any user-visible behavior changes
- known limitations that still apply

Avoid overstating maturity or production readiness.

## Post-Release Checklist

After publishing the release:

- [ ] confirm the tag exists on the remote
- [ ] confirm the GitHub release page is published
- [ ] confirm release notes are readable and accurate
- [ ] confirm the tagged commit matches the intended repository state
- [ ] confirm no accidental generated files were included

## Known Limitations for Current Releases

Current BlockGo releases still intentionally have these limitations:

- fixed validator set
- static peer configuration
- minimal HTTP API
- no smart contracts or VM
- no dynamic validator membership
- no advanced fork handling
- no hostile-network protections
- educational focus over production hardening

These limitations are acceptable for the current project goals and should remain visible in release communication.

## Future Improvements

Reasonable future release-process improvements include:

- automated GitHub release workflow on version tags
- prebuilt binary artifacts
- checksum generation
- release note automation
- container image publishing
- a dedicated changelog

Any automation added later should remain minimal and easy to understand.

## Summary

A good BlockGo release is:

- validated
- documented
- tagged clearly
- honest about scope
- easy for contributors to reproduce

Keep the process small, explicit, and maintainable.
