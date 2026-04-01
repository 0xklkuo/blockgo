# BlockGo

A minimal, open-source, educational blockchain project in Go.

## Status

**Milestone 0**: project foundation and OSS scaffolding.

No blockchain logic is implemented yet. This milestone exists to lock decisions early and keep the codebase simple, teachable, and contributor-friendly.

## Goals

- Teach core Go ecosystem practices through a real project
- Teach blockchain fundamentals by building a minimal serious prototype
- Stay small, readable, and easy to contribute to
- Prefer explicit design over framework-heavy abstraction
- Use real node-to-node networking, but keep the protocol understandable

## Non-Goals for V1

- Proof of Work
- full Proof of Stake
- smart contracts / VM
- Merkle Patricia Trie
- dynamic peer discovery
- advanced production hardening

## Locked V1 Decisions

- **Ledger model:** UTXO
- **Consensus:** Proof of Authority (fixed validator set, round-robin proposer)
- **P2P:** custom minimal TCP protocol with static peers
- **Storage:** bbolt
- **Crypto:** ed25519 + sha256
- **Approach:** stdlib-first, minimal dependencies

## Why This Design

This project is intentionally **not** a clone of Bitcoin or Ethereum internals.

Instead, it aims to teach:
- block structure
- transaction validation
- UTXO spending rules
- signed blocks
- peer synchronization
- persistence
- clean Go project structure

without dragging in complexity that would obscure the fundamentals.

## Quick Start

### Requirements
- Go 1.24.10
- Make

### Build
```bash
make build
```

### Test
```bash
make test
```

### Run CLI Stub
```bash
make run-cli
```

### Run Node Stub
```bash
make run-node
```

## Project Layout

```text
cmd/           Executables
internal/      Private implementation packages
configs/       Example config files
docs/          Architecture and design notes
test/          Integration test area
```

## Documentation

- [Architecture](./docs/architecture.md)
- [Config format](./docs/config.md)
- [Contributing](./CONTRIBUTING.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)

## Milestone Roadmap

- **M0**: foundation, docs, CI, repository scaffold
- **M1**: crypto primitives, tx/block types, hashing, merkle root, genesis
- **M2**: UTXO validation engine
- **M3**: persistence with bbolt
- **M4**: node loop, mempool, block production, PoA rules
- **M5**: P2P sync
- **M6**: CLI, optional HTTP API, Docker Compose
- **M7**: OSS polish and integration tests

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

## License

MIT
