# Contributing to BlockGo

Thanks for contributing to BlockGo.

BlockGo is a small, educational blockchain project in Go. The goal is to keep the codebase easy to read, easy to run, and easy to improve. Please optimize for simplicity, clarity, correctness, and contributor friendliness.

## Project Principles

When contributing, prefer:

- simple designs over clever designs
- explicit control flow over abstraction-heavy patterns
- small, focused packages
- standard library solutions unless a dependency clearly reduces complexity
- readable code over premature optimization
- minimal public surface area
- comments that improve learning value
- tests for core behavior and edge cases
- documentation updates when behavior changes

When in doubt, choose the smaller and clearer change.

## Scope and Philosophy

This project is intentionally minimal. Please avoid introducing complexity that does not clearly improve the educational value or the core blockchain implementation.

Good contributions usually:

- improve correctness
- improve readability
- improve tests
- improve docs
- fix inconsistencies between code, config, and docs
- keep the local developer workflow simple

Changes that should be discussed before implementation include:

- major architectural changes
- new dependencies
- framework adoption
- large API surface changes
- consensus model changes
- storage model changes
- networking protocol redesign

For non-trivial design changes, open an issue first.

## Development Setup

Requirements:

- Go 1.24.10
- Make
- Docker and Docker Compose for the local multi-node demo

Common commands:

```bash
make fmt
make vet
make test
make build
make ci
```

Useful local commands:

```bash
make run-cli
make run-node
go run ./cmd/blockgo gen-localnet -out ./configs/local
docker compose up --build
```

## Coding Guidelines

Please follow these guidelines:

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Keep packages small and focused
- Prefer concrete types until an interface is clearly needed
- Avoid premature abstraction
- Avoid unnecessary indirection
- Keep blockchain rules deterministic and explicit
- Keep transport, storage, and domain logic separated
- Use plain, readable names
- Keep error messages clear and actionable
- Do not introduce frameworks

### Go Style

- Run `make fmt` before submitting changes
- Run `make vet` and `make test`
- Keep functions short when practical
- Extract constants when they improve readability, not just to remove literals
- Prefer straightforward control flow over dense helper layers
- Keep comments accurate and up to date

### Blockchain-Specific Expectations

- Preserve deterministic behavior
- Be careful with validation logic and state transitions
- Prefer explicit validation over implicit assumptions
- Treat serialization and hashing changes as high-risk
- Keep consensus rules simple and well documented
- Add or update tests for transaction, block, state, storage, or sync behavior when relevant

## Documentation Expectations

If your change affects behavior, update the relevant documentation in the same pull request.

This may include:

- `README.md`
- `docs/architecture.md`
- `docs/config.md`
- config example files in `configs/`
- command help text
- Docker or Compose usage notes
- CI or release workflow docs

Docs should match the codebase exactly. Please avoid stale milestone text, outdated file paths, or commands that no longer work.

## Testing Expectations

Before opening a pull request, run:

```bash
make fmt
make vet
make test
make build
```

If your change affects repository-wide quality checks, also run:

```bash
make ci
```

Please include tests when changing:

- transaction validation
- block validation
- UTXO behavior
- persistence behavior
- consensus behavior
- P2P message handling
- node startup or sync behavior
- HTTP API request handling

If you cannot add a test for a valid reason, explain why in the pull request.

## Commit Style

Use conventional commits where practical. Examples:

- `feat: add block sync response validation`
- `fix: reject invalid transaction signatures`
- `docs: align README with final localnet workflow`
- `test: cover duplicate mempool submission`
- `chore: tighten release workflow`

Keep commits focused and easy to review.

## Pull Requests

Please keep pull requests small when possible.

A good pull request includes:

- a short summary of what changed
- the reason for the change
- notes on implementation choices if they are not obvious
- how the change was tested
- any follow-up work or known limitations

Before submitting, check that:

- code is formatted
- tests pass
- build passes
- docs are updated if needed
- config examples are updated if needed
- CI expectations still match the repository
- the change keeps the project minimal and readable

## Release Readiness

This repository aims to stay both GitHub-ready and release-ready.

When making changes, please avoid:

- placeholder documentation that contradicts the code
- dead commands in docs
- stale config paths
- unnecessary dependencies
- hidden behavior or surprising defaults
- broad refactors without a clear payoff

Release-oriented changes should preserve:

- reproducible local setup
- clear version output
- working build commands
- accurate docs
- consistent examples
- contributor-friendly workflows

## Security and Secrets

Please do not commit:

- private keys
- generated local demo secrets
- local data directories
- machine-specific credentials
- environment-specific secrets

Local demo configs that contain generated validator keys are for development only and should not be committed.

If you notice a security issue, avoid posting sensitive details publicly in an issue. Share a minimal report and maintainers can coordinate next steps.

## Questions and Design Discussion

If you are unsure whether a change fits the project goals, open an issue or draft pull request first.

Discussion is especially helpful for:

- architecture changes
- protocol changes
- storage changes
- release workflow changes
- contributor workflow changes

## Code of Conduct

By participating in this project, you agree to follow the [Code of Conduct](./CODE_OF_CONDUCT.md).

## License

By contributing, you agree that your contributions will be licensed under the project license.
