# BlockGo Architecture

## Purpose

BlockGo is a minimal educational blockchain implementation in Go.

The architecture is intentionally optimized for:

1. simplicity
2. correctness
3. readability
4. explicit control flow
5. minimal dependencies
6. contributor friendliness
7. realistic multi-node behavior without unnecessary abstraction

When trade-offs appear, the simpler and more teachable design should win unless it would make the system misleading or incorrect.

For protocol rules, config shape, and API contracts, see [`docs/spec.md`](./spec.md).

## System Overview

The current system includes:

- a UTXO-based ledger
- signed transactions and signed blocks
- a deterministic Proof of Authority proposer schedule
- a mempool for pending transactions
- persistent storage with bbolt
- a minimal TCP-based P2P protocol
- block synchronization between static peers
- a CLI for key generation, address derivation, transaction creation, and local demo setup
- an optional HTTP API for health, chain head, mempool, and transaction submission
- Docker and Docker Compose support for a local multi-node demo

This is not a production blockchain. It is a compact teaching system that shows the core moving parts clearly.

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

## Package Responsibilities

### `internal/blockchain`

Owns the domain model and ledger rules.

Responsibilities:

- transaction types
- block types
- hashing and Merkle root calculation
- genesis construction
- block validation
- transaction validation
- UTXO set updates

This package should stay independent from transport and storage concerns.

### `internal/consensus`

Owns Proof of Authority rules.

Responsibilities:

- validator set construction
- proposer selection by height
- block signing
- block signature validation

### `internal/crypto`

Owns small cryptographic helpers.

Responsibilities:

- key generation
- signing and signature verification
- address derivation from public keys

### `internal/mempool`

Owns in-memory pending transaction storage.

Responsibilities:

- add transactions
- reject duplicates
- remove included transactions
- expose a snapshot for block production

### `internal/node`

Owns runtime orchestration.

Responsibilities:

- startup and shutdown
- state initialization from genesis or storage
- mempool ownership
- block production
- external block application
- interaction with consensus, storage, and networking

This is the coordination layer of the system.

### `internal/p2p`

Owns peer-to-peer message transport.

Responsibilities:

- peer connections
- message encoding and decoding
- hello handshake
- block sync requests
- block propagation
- transaction propagation

### `internal/store`

Owns persistence through bbolt.

Responsibilities:

- save and load blocks
- save and load the current head
- save and load the UTXO set
- initialize persistent state from genesis

### `internal/config`

Owns configuration loading and validation.

Responsibilities:

- node config parsing
- genesis config parsing
- relative path resolution
- startup validation

### `internal/api`

Owns the optional HTTP transport surface.

Responsibilities:

- HTTP routing
- request decoding
- response encoding
- mapping HTTP requests onto node behavior

The API should remain thin. Blockchain rules should stay outside this package.

### `internal/version`

Owns build and version metadata for the binaries.

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

## Runtime Boundaries

The intended dependency boundaries are:

- `blockchain` owns ledger rules
- `consensus` owns proposer and signature rules
- `node` coordinates runtime behavior
- `p2p` transports peer messages
- `api` exposes HTTP endpoints
- `store` persists state
- `config` loads startup inputs

Guidelines:

- do not move ledger rules into `api`, `p2p`, or `config`
- keep consensus logic separate from storage and transport
- keep `node` as an orchestrator, not a place for ad hoc rule duplication
- prefer concrete dependencies until an interface is clearly needed
- add abstractions only when they reduce complexity rather than hide it

## Documentation Boundary Decision

As part of the documentation refactor, the project now keeps architecture, specification, and roadmap content in separate docs:

- `README.md` for entry-point guidance
- `docs/spec.md` for rules and contracts
- `docs/architecture.md` for codebase structure and runtime flow
- `docs/roadmap.md` for direction and deferred work

This mirrors the intended package boundaries and reduces duplicated guidance across the repository.
