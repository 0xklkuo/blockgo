# Contributing to BlockGo

Thanks for contributing.

## Project Principles

Please optimize for:

- simplicity
- clarity
- correctness
- small APIs
- explicit control flow
- low dependency count
- educational readability

When in doubt, prefer the simpler design.

## Development Setup

Requirements:
- Go 1.24.10
- Make

Common commands:

```bash
make fmt
make vet
make test
make build
make ci
```

## Coding Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Keep packages small and focused
- Prefer the standard library unless a dependency clearly reduces complexity
- Avoid premature interfaces
- Add comments where they improve learning value
- Keep business logic decoupled from transport and storage
- Do not introduce frameworks

## Commit Style

Use conventional commits where practical, for example:

- `feat: add transaction hashing`
- `fix: reject duplicate utxo spends in block`
- `docs: clarify p2p handshake`
- `chore: update ci workflow`

## Pull Requests

Please include:

- what changed
- why it changed
- how it was tested
- any follow-up work

Small PRs are preferred.

## Documentation Expectations

If behavior changes, update the relevant docs:
- README
- docs/
- config examples
- command help text

## Testing Expectations

At minimum, before opening a PR:

```bash
make ci
```

As the project grows, new core logic should include tests.

## Questions and Design Changes

If you want to introduce a non-trivial architectural change, open an issue first to discuss it.
