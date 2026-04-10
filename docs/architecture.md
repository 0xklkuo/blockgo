# BlockGo Architecture

## Purpose

BlockGo is a minimal, educational blockchain implementation in Go.

It is designed to help developers learn the fundamentals of:

- blockchain data structures
- transaction validation
- UTXO-based state
- Proof of Authority consensus
- peer-to-peer synchronization
- persistence
- clean Go project structure

The project intentionally favors clarity over feature breadth. It aims to be small enough to read end to end, while still being realistic enough to demonstrate how a simple blockchain node works.

## Design Priorities

BlockGo is built around these priorities:

1. Simplicity
2. Correctness
3. Readability
4. Minimal dependencies
5. Explicit control flow
6. Contributor friendliness
7. Real multi-node behavior without unnecessary abstraction

When trade-offs appear, the simpler and more teachable design should win unless it would make the system misleading or incorrect.

## Implemented System Overview

The current system includes:

- a UTXO-based ledger
- signed transactions
- signed blocks
- a deterministic Proof of Authority proposer schedule
- a mempool for pending transactions
- persistent storage with bbolt
- a minimal TCP-based P2P protocol
- block synchronization between static peers
- a CLI for key generation, address derivation, transaction creation, and local demo setup
- an optional HTTP API for health, chain head, mempool, and transaction submission
- Docker and Docker Compose support for a local multi-node demo

This is not a production blockchain. It is a compact teaching system that demonstrates the core moving parts clearly.

## High-Level Architecture

```text
                +----------------------+
                |      blockgo CLI     |
                | keygen, address, tx  |
                | localnet generation  |
                +----------+-----------+
                           |
                           v
+--------------------------------------------------------------+
|                         blockgo-node                         |
|                                                              |
|  +----------------+   +----------------+   +---------------+ |
|  |    config      |   |      api       |   |      p2p      | |
|  | JSON loading   |   | optional HTTP  |   | TCP protocol  | |
|  | validation     |   | endpoints      |   | peer sync     | |
|  +--------+-------+   +--------+-------+   +-------+-------+ |
|           |                    |                   |         |
|           +--------------------+-------------------+         |
|                                v                             |
|                      +-------------------+                   |
|                      |       node        |                   |
|                      | lifecycle         |                   |
|                      | mempool           |                   |
|                      | block production  |                   |
|                      | block acceptance  |                   |
|                      +----+---------+----+                   |
|                           |         |                        |
|                           v         v                        |
|                 +-------------+   +----------------+         |
|                 | blockchain  |   |   consensus    |         |
|                 | tx/block    |   | PoA validator  |         |
|                 | validation  |   | proposer rules |         |
|                 | UTXO state  |   | block signing  |         |
|                 +------+------+   +--------+-------+         |
|                        |                   |                 |
|                        +---------+---------+                 |
|                                  v                           |
|                           +-------------+                    |
|                           |    store    |                    |
|                           | bbolt DB    |                    |
|                           | blocks      |                    |
|                           | UTXO set    |                    |
|                           +-------------+                    |
+--------------------------------------------------------------+
```

## Core Packages

### `internal/blockchain`

This package contains the core domain model and validation logic.

Responsibilities:

- transaction types
- block types
- transaction hashing
- block hashing
- Merkle root calculation
- genesis block construction
- block validation
- transaction validation
- UTXO set updates

This package is the heart of the ledger rules. It should remain independent from transport and storage concerns.

### `internal/consensus`

This package implements the minimal Proof of Authority rules.

Responsibilities:

- validator set construction
- proposer selection by block height
- block signing
- block signature validation

The current proposer schedule is deterministic and round-robin across a fixed validator set.

### `internal/crypto`

This package contains small cryptographic helpers.

Responsibilities:

- ed25519 key generation
- signing and signature verification
- address derivation from public keys

The project uses:

- `ed25519` for signatures
- `sha256` for transaction and block hashing

### `internal/mempool`

This package stores pending transactions before they are included in a block.

Responsibilities:

- add transactions
- reject duplicates
- remove included transactions
- expose a snapshot for block production

The mempool is intentionally simple and in-memory.

### `internal/node`

This package coordinates the running node.

Responsibilities:

- startup and shutdown
- state initialization from genesis or storage
- mempool ownership
- block production
- external block application
- P2P message handling
- interaction with storage and consensus

This package is the orchestration layer of the system.

### `internal/p2p`

This package implements the minimal peer-to-peer protocol over TCP.

Responsibilities:

- peer connections
- message encoding and decoding
- hello handshake
- block sync requests
- block propagation
- transaction propagation

The protocol is intentionally small and readable.

### `internal/store`

This package persists chain state using bbolt.

Responsibilities:

- save and load blocks
- save and load the current head
- save and load the UTXO set
- initialize persistent state from genesis

Persistence is kept simple so contributors can understand exactly what is stored and when.

### `internal/config`

This package loads and validates JSON configuration.

Responsibilities:

- node config parsing
- genesis config parsing
- relative path resolution
- startup validation

Configuration is strict and explicit to keep node startup predictable.

### `internal/api`

This package exposes a small optional HTTP API.

Current endpoints:

- `GET /healthz`
- `GET /v1/chain/head`
- `GET /v1/mempool`
- `POST /v1/transactions`

The API is intentionally thin. It should remain a transport layer over the node, not a place for blockchain business logic.

### `internal/version`

This package exposes build and version metadata for the binaries.

## Ledger Model

BlockGo uses a UTXO ledger.

Why this model was chosen:

- it makes ownership and spending rules explicit
- it teaches transaction validation clearly
- it keeps state transitions easy to reason about
- it fits a minimal educational blockchain well

Each non-coinbase transaction:

- references previous outputs
- proves ownership with a public key and signature
- creates new outputs
- must not overspend inputs
- must not double-spend within the same transaction or block

The UTXO set is updated as blocks are applied.

## Transaction Model

A transaction contains:

- an ID
- zero or more inputs
- one or more outputs

### Inputs

Each input contains:

- a reference to a previous output
- the spender public key
- a signature

### Outputs

Each output contains:

- a value
- a destination address

### Coinbase Transactions

A coinbase transaction has:

- no inputs
- one or more outputs

In the current design, each produced block includes a coinbase transaction that pays:

- the block reward
- the total fees from included transactions

## Block Model

A block contains:

- a header
- a list of transactions
- a validator signature
- a block hash

The header includes:

- height
- previous block hash
- Merkle root
- timestamp
- validator public key

A block is finalized by:

1. finalizing any transactions that do not yet have IDs
2. computing the Merkle root
3. hashing the header payload
4. signing the block with the validator key

## Consensus Model

BlockGo uses a minimal Proof of Authority model.

### Current Rules

- the validator set is fixed in config/genesis for the current system shape
- validators are identified by public keys
- the proposer for a height is selected deterministically
- only the expected proposer may produce a valid block for that height
- blocks must be signed by the proposer

### Why PoA

PoA was chosen because it is:

- simple to explain
- efficient for local multi-node demos
- much easier to implement correctly than Proof of Work or full Proof of Stake
- sufficient for teaching block production and validation flow

### What It Does Not Try to Solve

This project does not currently implement:

- validator set changes
- slashing
- stake economics
- Byzantine finality gadgets
- advanced fork choice rules

## Node Lifecycle

At a high level, a node starts like this:

1. load node config
2. load genesis config
3. validate config and genesis
4. open the local store
5. load persisted head and UTXO state, or initialize from genesis
6. start the P2P server
7. connect to configured peers
8. start the block production loop
9. optionally start the HTTP API

During runtime, the node:

- accepts transactions into the mempool
- validates transactions before mempool insertion
- periodically checks whether it is the expected proposer
- builds and signs a block when eligible
- validates and persists produced blocks
- broadcasts new transactions and blocks
- accepts and validates blocks from peers
- updates the local head and UTXO set

## P2P Protocol

The P2P layer is intentionally minimal.

### Current Message Types

- `hello`
- `get_blocks`
- `blocks`
- `new_tx`
- `new_block`

### Sync Flow

A simplified sync flow looks like this:

1. a peer connects
2. peers exchange `hello` messages with current height
3. if one peer is behind, it requests blocks starting from the next height
4. the other peer responds with `blocks`
5. the receiving node validates and applies each block in order

This is intentionally simple and suitable for a small linear chain demo.

## Persistence Model

The node persists enough state to restart cleanly.

Persisted state includes:

- blocks
- current head
- UTXO set

This allows the node to recover its latest known chain state without rebuilding everything from scratch on each startup.

The storage layer is intentionally direct and compact. It is not designed for high-throughput production workloads.

## Configuration Model

BlockGo uses JSON configuration files.

### Node Config

The node config defines:

- node ID
- data directory
- P2P listen address
- optional HTTP address
- static peer list
- genesis file path
- block interval
- max transactions per block
- local validator private key
- validator public keys

### Genesis Config

The genesis config defines:

- chain ID
- creation timestamp metadata
- block interval metadata
- validator set
- initial allocations

The runtime genesis block is built from the config rather than loaded as a precomputed binary artifact.

## CLI and Developer Experience

The CLI currently supports:

- version output
- key generation
- address derivation
- transaction creation
- local network config generation

The local network generator creates a runnable three-node demo configuration with:

- validator keys
- node configs
- a shared genesis file

This keeps the onboarding path short for new contributors.

## HTTP API

The optional HTTP API is intentionally small.

### Goals

- provide a simple inspection surface
- support local demos and manual testing
- avoid mixing transport concerns into core blockchain logic

### Current Endpoints

- `GET /healthz` returns a basic health response
- `GET /v1/chain/head` returns the current head block
- `GET /v1/mempool` returns the current mempool size
- `POST /v1/transactions` accepts a signed transaction payload

This API is not intended to be a complete wallet or explorer API.

## Docker and Local Demo

The repository includes Docker support and a Compose-based local demo.

The intended local demo flow is:

1. generate local config files
2. start three nodes with Docker Compose
3. inspect health endpoints
4. submit transactions and observe block production

This gives contributors a realistic but still understandable multi-node environment.

## Boundaries and Dependency Rules

The intended package boundaries are:

- `blockchain` owns ledger rules
- `consensus` owns proposer and signature rules
- `node` coordinates runtime behavior
- `p2p` transports messages
- `api` exposes HTTP endpoints
- `store` persists state
- `config` loads and validates configuration

Important design rule:

- transport layers should not own business logic
- storage should not define consensus rules
- core validation should remain testable without network or HTTP dependencies

## Non-Goals

BlockGo intentionally does not include the following in its current scope:

- Proof of Work
- full Proof of Stake
- smart contracts or a VM
- account-based state
- Merkle Patricia Trie
- dynamic peer discovery
- dynamic validator rotation
- advanced fork handling
- production-grade networking hardening
- production-grade observability
- production-grade security hardening

These are excluded to keep the project focused and teachable.

## Known Simplifications

The current implementation makes several deliberate simplifications:

- static peer configuration
- fixed validator set
- simple linear sync assumptions
- in-memory mempool
- minimal HTTP API
- no wallet subsystem
- no transaction pool prioritization beyond simple selection
- no advanced chain reorganization logic

These are acceptable for the educational goals of the project.

## Growth Path

Reasonable future extensions, if kept minimal, could include:

- richer integration tests
- better API documentation
- chain inspection CLI commands
- improved transaction submission tooling
- clearer observability and structured logs
- limited fork-handling improvements
- release automation

Any future work should preserve the project's core values:

- minimalism
- clarity
- correctness
- educational readability

## Summary

BlockGo is a compact blockchain node built to teach the essentials well.

It demonstrates how a minimal blockchain system can be structured in Go with:

- explicit domain logic
- small packages
- simple configuration
- deterministic consensus
- understandable networking
- persistent state
- contributor-friendly project boundaries

The architecture is intentionally modest. That is a feature, not a limitation.
