# BlockGo

A minimal, open-source, educational blockchain project in Go.

BlockGo is built to help developers learn the fundamentals of the Go ecosystem and modern blockchain development by implementing a small but real blockchain from scratch. The project favors explicit design, small packages, minimal dependencies, and contributor-friendly code over framework-heavy abstraction or production-scale complexity.

## Status

**Milestone 7**: OSS polish and release-ready repository.

The codebase now includes:

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
- OSS contributor scaffolding and CI

## Goals

- Teach core Go ecosystem practices through a real project
- Teach blockchain fundamentals by building a minimal serious prototype
- Stay small, readable, and easy to contribute to
- Prefer explicit design over framework-heavy abstraction
- Use real node-to-node networking, but keep the protocol understandable

## Non-Goals for MVP

- Proof of Work
- full Proof of Stake
- smart contracts / VM
- Merkle Patricia Trie
- dynamic peer discovery
- dynamic validator set changes
- advanced production hardening
- anti-abuse and adversarial networking defenses

## Locked MVP Decisions

- **Ledger model:** UTXO
- **Consensus:** Proof of Authority (fixed validator set, round-robin proposer)
- **P2P:** custom minimal TCP protocol with static peers
- **Storage:** bbolt
- **Crypto:** ed25519 + sha256
- **Approach:** stdlib-first, minimal dependencies

## Why This Design

This project is intentionally not a clone of Bitcoin or Ethereum internals.

Instead, it aims to teach:

- block structure
- transaction validation
- UTXO spending rules
- signed blocks
- peer synchronization
- persistence
- clean Go project structure

Without dragging in complexity that would obscure the fundamentals.

## Requirements

- Go 1.24.10
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
blockgo-node -version
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

```bash
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

If you need to change the output:

```bash
./bin/blockgo create-tx -in-prev-tx <tx-id-hex> -in-index 0 -in-pubkey <public-key-hex> -in-privkey <private-key-hex> -out-address <recipient-address-hex> -out-value 100 -change-address <change-address-hex> -change-value 900
```

## Run a Single Node

Use the example config files as a starting point:

- `configs/genesis.example.json`
- `configs/node.example.json`

Run the node:

```bash
./bin/blockgo-node -config ./configs/node.example.json
```

## Local Multi-Node Demo

This repository includes a simple 3-node local demo using Docker Compose.

### 1. Generate Local Network Config

Generate a runnable local network with validator keys:

```bash
go run ./cmd/blockgo gen-localnet -out ./configs/local
```

This creates:

- `configs/local/genesis.json`
- `configs/local/node1.json`
- `configs/local/node2.json`
- `configs/local/node3.json`

These generated files contain local demo private keys and should not be committed.

### 2. Start the Network

```bash
docker compose up --build
```

Refer to the official Docker Compose documentation for more details: https://docs.docker.com/compose/

### 3. Check Node Health

```bash
curl http://localhost:8001/healthz
curl http://localhost:8002/healthz
curl http://localhost:8003/healthz
```

### 4. Inspect Chain Head

```bash
curl http://localhost:8001/v1/chain/head
```

### 5. Inspect Mempool Size

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

Example request:

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

## Project Layout

```text
cmd/           Executables
internal/      Private implementation packages
configs/       Example config files
docs/          Architecture and design notes
scripts/       Helper scripts
test/          Integration test area
bin/           Local build output
.github/       OSS templates and CI
```

## Documentation

- [Architecture](./docs/architecture.md)
- [Config format](./docs/config.md)
- [Release guide](./docs/release.md)
- [Contributing](./CONTRIBUTING.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)

## Development Workflow

Common commands:

```bash
make fmt
make vet
make test
make build
make ci
```

## Release Readiness Notes

This repository is release-ready for an educational `v0.x` style release, with the following scope:

- minimal blockchain implementation
- educational clarity over production hardening
- fixed validator set
- static peer configuration
- no smart contract runtime
- no dynamic consensus membership
- no hostile-network protections

### Tagging and Releases

Use semantic version tags for releases, for example:

```bash
git tag v0.1.0
git push origin v0.1.0
```

For the full release workflow, release checklist, and versioning guidance, see [docs/release.md](./docs/release.md).

## Milestone Roadmap

- **M0**: foundation, docs, CI, repository scaffold
- **M1**: crypto primitives, tx/block types, hashing, merkle root, genesis
- **M2**: UTXO validation engine
- **M3**: persistence with bbolt
- **M4**: node loop, mempool, block production, PoA rules
- **M5**: P2P sync
- **M6**: CLI, optional HTTP API, Docker Compose
- **M7**: OSS polish, integration alignment, release-ready repository

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

If behavior changes, please also update the relevant docs and examples.

## License

MIT
