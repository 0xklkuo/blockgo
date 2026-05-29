# BlockGo Roadmap

## Purpose

This document captures the current project direction, completed milestones, release posture, and the work intentionally deferred.

It is not a promise of exact delivery dates. It is a guide for keeping the project coherent, small, and maintainable.

## Current Position

BlockGo is at an educational `v0.x` stage.

The repository is already functional and validated around a compact feature set:

- UTXO ledger and validation
- Proof of Authority with a fixed validator set
- bbolt persistence
- minimal TCP peer sync
- CLI utilities
- optional HTTP API
- Docker Compose local multi-node demo
- CI and release automation

The project should continue to improve clarity and maintainability without losing its small-scope teaching value.

## Documentation Strategy Decision

This refactor intentionally consolidates project guidance into four core docs:

- `README.md`
- `docs/spec.md`
- `docs/architecture.md`
- `docs/roadmap.md`

Rationale:

- reduce duplicated guidance
- make the repository easier to scan
- keep onboarding, behavior, structure, and direction separate
- lower the chance of stale or contradictory docs

As part of this decision, durable guidance from the old standalone docs was merged into the core set:

- `docs/config.md` → `docs/spec.md`
- `docs/release.md` → `README.md` and `docs/roadmap.md`
- `CONTRIBUTING.md` → `README.md` and `docs/roadmap.md`
- `CODE_OF_CONDUCT.md` → `README.md`

Future doc work should preserve this structure unless there is a strong reason to expand it. Transient release notes and one-off outcome writeups should live in GitHub releases, pull requests, or issues rather than long-lived repository docs.

## Completed Milestones

- **M0**: foundation, docs, CI, repository scaffold
- **M1**: crypto primitives, transaction and block types, hashing, Merkle root, genesis
- **M2**: UTXO validation engine
- **M3**: persistence with bbolt
- **M4**: node loop, mempool, block production, PoA rules
- **M5**: P2P sync
- **M6**: CLI, optional HTTP API, Docker Compose
- **M7**: OSS polish, integration alignment, release-ready repository

## Refactor Continuation Plan

This refactor is intentionally split into small milestones so a future agent can continue the work without reconstructing prior context.

### Refactor Milestone 1 — documentation consolidation

Status: completed.

Scope:

- keep the repository centered on `README.md`, `docs/spec.md`, `docs/architecture.md`, and `docs/roadmap.md`
- migrate durable guidance out of redundant standalone docs
- remove stale or duplicated long-lived docs
- align the README with actual repository behavior

Completion notes:

- removed `docs/config.md`, `docs/release.md`, `CONTRIBUTING.md`, and `CODE_OF_CONDUCT.md`
- moved durable contribution, conduct, and release guidance into the core docs
- clarified that tracked example node configs are templates, not runnable as committed

Validation target:

- diagnostics clean
- `make fmt`
- `make ci`

### Refactor Milestone 2 — runtime hardening and local developer flow

Status: completed.

Scope:

- make `make run-node` work without manual placeholder editing
- keep the Docker localnet flow working as documented
- add focused tests for the local config generation path
- harden HTTP server defaults with explicit timeouts
- harden `POST /v1/transactions` request decoding
- add focused API tests for accepted and rejected request shapes

Acceptance criteria:

- `make run-node` starts from a generated local config instead of failing on placeholder values
- Docker-oriented config generation still works for the Compose demo
- malformed, oversized, or unknown-field transaction requests are rejected predictably
- `make test` and `make ci` pass

Completion notes:

- `make run-node` now generates an isolated single-node local config in `configs/run-node`
- local config generation supports `-mode` and `-nodes` so local and Docker flows stay explicit
- HTTP server defaults now include explicit timeouts
- the transaction submission endpoint now rejects oversized, trailing, or unknown-field JSON payloads
- focused tests cover local config generation and API request decoding behavior

Estimated effort:

- 2 to 4 senior-engineer hours

### Refactor Milestone 3 — CI and release workflow cleanup

Status: completed.

Scope:

- reduce redundant validation work in the release workflow
- keep local and CI build flags aligned where practical
- review release discoverability and final documentation polish
- do a final pass on repository clarity, maintainability, and agent handoff quality

Acceptance criteria:

- release automation is easier to understand and avoids obvious duplicated work
- docs still match the final workflow
- `make ci` remains the main local validation command

Completion notes:

- tag pushes now rely on the dedicated release workflow instead of also running the branch CI workflow
- the release workflow validates once, then fans out to cross-build and package artifacts
- release build flags now come from `Makefile` through the `release-dist` target instead of being duplicated in workflow YAML
- roadmap and spec docs were aligned with the final local and release workflows

Estimated effort:

- 1 to 2 senior-engineer hours

## Near-Term Improvements

Good next improvements for the current scope include:

- keep focused tests growing only where externally visible behavior changes
- improve observability and debugging guidance without bloating the runtime
- keep config examples, CLI help text, and docs aligned as the project evolves
- preserve a clean local developer workflow

These improvements should remain incremental and low-risk.

## Longer-Term Possibilities

Reasonable future work, if it still serves the educational goals:

- clearer observability and diagnostics
- more explicit sync behavior documentation and tests
- compatibility checks for release artifacts
- modest API improvements for demo usability
- more end-to-end examples around transaction lifecycle

These should be added only if they improve teaching value without bloating the system.

## Explicit Non-Goals

Unless the project goals change, the following remain out of scope:

- Proof of Work
- full Proof of Stake
- smart contracts or a VM
- dynamic validator membership
- dynamic peer discovery
- hostile-public-network hardening
- complex fork choice or finality protocols
- framework-heavy rewrites

## Release Direction

BlockGo releases should stay honest about the current maturity level.

Current release posture:

- educational `v0.x` scope
- semantic version tags such as `v0.1.0`
- GitHub releases are published from `v*` tags through the release workflow
- the release workflow validates once, then cross-builds and publishes packaged artifacts
- release notes should describe actual repository state only
- documentation and example configs should match the codebase

A release is in good shape when:

- `make ci` passes
- docs are aligned with behavior
- example configs still match the implementation
- the local demo flow is still understandable and reproducible

## Change Guidance

Good changes usually:

- improve correctness
- improve readability
- improve tests
- improve docs
- simplify workflows
- reduce duplicated or misleading guidance

Changes worth discussing before implementation include:

- major architectural changes
- new dependencies
- consensus changes
- storage model changes
- networking protocol redesign
- broad refactors without a clear payoff

## Acceptance Standard for Future Milestones

A good milestone should leave the repository:

- simpler than before
- validated by the relevant checks or tests
- documented where behavior changed
- easy for the next contributor or agent to understand

That standard matters more than adding more features.
