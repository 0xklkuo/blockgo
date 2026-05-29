[![CI](https://github.com/0xklkuo/blockgo/actions/workflows/ci.yml/badge.svg)](https://github.com/0xklkuo/blockgo/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/0xklkuo/blockgo?label=release&sort=semver)](https://github.com/0xklkuo/blockgo/releases) [![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE) [![Go version](https://img.shields.io/badge/go-1.24.10-blue.svg)](https://go.dev/)

# BlockGo

BlockGo is a minimal, open-source, educational blockchain project in Go.

It is designed to teach core blockchain and Go concepts through a small but real implementation: UTXO state, signed transactions, signed blocks, fixed-validator Proof of Authority, local persistence, basic peer sync, a tiny HTTP API, and a runnable multi-node demo.

## Goals

- keep the codebase small, readable, and teachable
- prefer explicit design over framework-heavy abstraction
- demonstrate realistic node-to-node behavior without hiding the moving parts
- keep dependencies minimal and contributor workflow simple

## Current Scope

BlockGo currently includes:

- core blockchain data structures
- transaction hashing and signing
- UTXO validation and state transitions
- genesis creation and validation
- bbolt-backed persistence
- Proof of Authority consensus with a fixed validator set
- minimal TCP peer-to-peer sync
- node runtime with block production
- CLI utilities
- optional HTTP API
- Docker Compose local multi-node demo
- CI and release automation

### Non-Goals

The current project intentionally does not include:

- Proof of Work
- full Proof of Stake
- smart contracts or a VM
- dynamic peer discovery
- dynamic validator set changes
- advanced fork choice or finality gadgets
- hostile-network hardening
- production-scale performance tuning

### Locked Design Decisions

- **Ledger model:** UTXO
- **Consensus:** Proof of Authority with a fixed validator set and deterministic proposer selection
- **P2P:** small custom TCP protocol with static peers
- **Storage:** bbolt
- **Crypto:** ed25519 for signatures and sha256 for hashing
- **Approach:** stdlib-first and minimal dependencies

For the exact behavior and contracts, see [`docs/spec.md`](./docs/spec.md).

## Requirements

- Go `1.24.10`
- Make
- Docker and Docker Compose for the local multi-node demo

## Quick Start

### Build

```bash
make build
```

### Test

```bash
make test
```

### Run All Local Checks

```bash
make ci
```

### Show CLI Version

```bash
make run-cli
```

### Show Node Version

```bash
./bin/blockgo-node -version
```

## CLI Overview

After building, the CLI binary is available at `./bin/blockgo`.

### Commands

- `blockgo version`
- `blockgo generate-key`
- `blockgo gen-key`
- `blockgo address -pubkey <hex>`
- `blockgo create-tx ...`
- `blockgo gen-localnet -out <dir>`

### Generate a Key Pair

```bash
./bin/blockgo generate-key
```

Example output:

```json
{
  "public_key_hex": "<hex>",
  "private_key_hex": "<hex>",
  "address": "<address-hex>"
}
```

### Derive an Address From a Public Key

```bash
./bin/blockgo address -pubkey <public-key-hex>
```

### Create a Signed Transaction

```bash
./bin/blockgo create-tx -in-prev-tx <tx-id-hex> -in-index 0 -in-pubkey <public-key-hex> -in-privkey <private-key-hex> -out-address <recipient-address-hex> -out-value 100
```

Example with a change output:

```bash
./bin/blockgo create-tx -in-prev-tx <tx-id-hex> -in-index 0 -in-pubkey <public-key-hex> -in-privkey <private-key-hex> -out-address <recipient-address-hex> -out-value 100 -change-address <change-address-hex> -change-value 900
```

## Run a Single Node

Use the tracked example configs as templates:

- `configs/genesis.example.json`
- `configs/node.example.json`

These files are not runnable as committed because they intentionally contain placeholder keys and addresses. Replace the placeholders with real values before using them directly.

Run the node after filling in the example config:

```bash
./bin/blockgo-node -config ./configs/node.example.json
```

If you want a runnable setup without manual key editing, use the generated local config flow in the multi-node demo and point the node at one of the generated files such as `configs/local/node1.json`.

## Local Multi-Node Demo

### 1. Generate local config

```bash
go run ./cmd/blockgo gen-localnet -out ./configs/local
```

This creates:

- `configs/local/genesis.json`
- `configs/local/node1.json`
- `configs/local/node2.json`
- `configs/local/node3.json`

These generated files contain local demo private keys and should not be committed.

### 2. Start the network

```bash
docker compose up --build
```

### 3. Check node health

```bash
curl http://localhost:8001/healthz
curl http://localhost:8002/healthz
curl http://localhost:8003/healthz
```

### 4. Inspect chain head

```bash
curl http://localhost:8001/v1/chain/head
```

### 5. Inspect mempool size

```bash
curl http://localhost:8001/v1/mempool
```

### Notes

- all nodes share the same generated `genesis.json`
- validator keys and node configs are generated automatically
- generated files are intended for local development and demo use only
- the local demo uses static peers and a fixed validator set

## HTTP API

The node exposes a small optional HTTP API.

### Endpoints

- `GET /healthz`
- `GET /v1/chain/head`
- `GET /v1/mempool`
- `POST /v1/transactions`

### Submit a Transaction

Example payload:

```json
{
  "inputs": [
    {
      "prev_tx_id": "<tx-id-hex>",
      "output_index": 0,
      "public_key_hex": "<public-key-hex>",
      "signature_hex": "<signature-hex>"
    }
  ],
  "outputs": [
    {
      "value": 100,
      "address": "<recipient-address-hex>"
    }
  ]
}
```

Example call:

```bash
curl -X POST http://localhost:8001/v1/transactions -H "Content-Type: application/json" --data @tx.json
```

See [`docs/spec.md`](./docs/spec.md) for the config and API contract details.

## Project Layout

```text
cmd/           Executables
internal/      Private implementation packages
configs/       Example config files
docs/          Core project documentation
scripts/       Helper scripts
test/          Integration test area
bin/           Local build output
.github/       OSS templates and CI
```

## Documentation Map

The project documentation is intentionally centered on four core docs:

- [`README.md`](./README.md): entry point, workflow, contribution guidance, and community expectations
- [`docs/spec.md`](./docs/spec.md): behavior, data contracts, config, and API surface
- [`docs/architecture.md`](./docs/architecture.md): package boundaries, runtime flow, and system structure
- [`docs/roadmap.md`](./docs/roadmap.md): current status, milestones, release direction, and deferred work

This documentation split is a deliberate refactor decision to keep the project easier to scan, easier to maintain, and less prone to duplicated or stale guidance. Durable material that used to live in standalone config, release, contributing, and conduct docs now lives in these four documents.

## Working Style and Contribution Expectations

Please keep changes:

- simple, explicit, and easy to review
- aligned with the educational scope of the project
- focused on correctness, readability, and maintainability
- accompanied by doc updates when behavior changes
- accompanied by tests when validation, state, storage, sync, or API behavior changes

Prefer:

- small, focused packages
- standard library solutions unless a dependency clearly reduces complexity
- straightforward control flow over abstraction-heavy patterns
- clear error messages and readable names
- deterministic blockchain rules and explicit validation

Discuss first before making major changes to:

- architecture
- consensus rules
- storage model
- networking protocol shape
- dependency strategy
- large API surface changes

### Pull Requests and Commits

Keep pull requests small when practical. A good pull request explains:

- what changed
- why the change was needed
- how it was validated
- any follow-up work or notable limitations

Use conventional commits where practical, for example:

- `feat: ...`
- `fix: ...`
- `docs: ...`
- `test: ...`
- `chore: ...`

### Community Expectations

Project spaces include issues, pull requests, discussions, and related collaboration channels.

Please:

- be respectful, constructive, and inclusive
- give actionable technical feedback
- focus on the code and ideas, not the person
- avoid harassment, personal attacks, discrimination, or deliberate disruption

Maintainers may edit, remove, or reject comments, issues, pull requests, or contributions that do not meet these expectations.

### Security Expectations

- do not commit private keys, generated demo secrets, local data directories, or machine-specific credentials
- generated local demo configs are for development only and should not be committed
- if you notice a security issue, share a minimal report and avoid posting sensitive details publicly

### Contribution License

By contributing to BlockGo, you agree that your contributions are licensed under the repository's MIT license.

## Development Workflow

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
go run ./cmd/blockgo gen-localnet -out ./configs/local
docker compose up --build
```

Use `make run-node` only after replacing the placeholder values in `configs/node.example.json`.

## License

MIT
