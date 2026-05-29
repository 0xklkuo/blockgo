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

## Near-Term Improvements

Good next improvements for the current scope include:

- tighten CI and release workflow efficiency
- harden HTTP API request handling and server defaults
- expand focused tests where behavior is externally visible
- keep config examples, CLI help text, and docs aligned
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
