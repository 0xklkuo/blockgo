# BlockGo Specification

## Purpose

This document defines the current behavior, data contracts, and user-facing technical rules of BlockGo.

If a question is about ledger rules, config shape, API payloads, or intentionally fixed design choices, this document is the source of truth.

## Scope and Fixed Decisions

BlockGo is intentionally a small educational blockchain.

The current system is built around these fixed decisions:

- **ledger model:** UTXO
- **consensus:** Proof of Authority with a fixed validator set
- **proposer selection:** deterministic and height-based
- **networking:** static peers over a minimal TCP protocol
- **storage:** bbolt
- **signatures:** ed25519
- **hashing:** sha256
- **project approach:** stdlib-first and minimal dependencies

These decisions are intentionally conservative. They keep the codebase readable and the behavior easy to reason about.

## Ledger Model

BlockGo uses a UTXO ledger.

Why:

- ownership and spending rules are explicit
- transaction validation is straightforward to teach
- state transitions are easy to reason about
- the model fits a small educational blockchain well

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

Each produced block includes a coinbase transaction that pays:

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

- the validator set is fixed in config and genesis for the current system shape
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

### Out of Scope

The current system does not implement:

- validator set changes
- slashing
- stake economics
- Byzantine finality gadgets
- advanced fork choice rules

## P2P Protocol

The peer-to-peer layer is intentionally minimal.

### Current Message Types

- `hello`
- `get_blocks`
- `blocks`
- `new_tx`
- `new_block`

### Sync Flow

A simplified sync flow is:

1. a peer connects
2. peers exchange `hello` messages with current height
3. if one peer is behind, it requests blocks starting from the next height
4. the other peer responds with `blocks`
5. the receiving node validates and applies each block in order

This protocol is intentionally small and suited to a simple linear-chain demo.

## Persistence Contract

The node persists enough state to restart cleanly.

Persisted state includes:

- blocks
- current head
- UTXO set

This allows the node to recover the latest known chain state without rebuilding everything from scratch on each startup.

## Configuration Contract

BlockGo uses JSON configuration files for both node startup and genesis state.

Tracked examples:

- `configs/node.example.json`
- `configs/genesis.example.json`

For generated local config files, use:

```bash
go run ./cmd/blockgo gen-localnet [-mode docker|local] [-nodes n] -out ./configs/local
```

That command creates:

- `configs/local/genesis.json`
- `configs/local/node1.json`
- `configs/local/node2.json`
- `configs/local/node3.json`

These generated files are for local development only and include demo private keys.

Generation modes:

- `docker` is the default and emits Compose-oriented addresses such as `node2:7002` and `/app/configs/local/genesis.json`
- `local` emits loopback addresses such as `127.0.0.1:7001`, an isolated data directory under `data/run-node/`, and a config-relative genesis path so the generated node config can be used directly for local development
- `-nodes` controls how many validator configs are emitted; `make run-node` uses `-nodes 1`, while the Docker demo uses `-nodes 3`

### Node Configuration

A node config defines how a single node starts.

Example shape:

```json
{
  "node_id": "node1",
  "data_dir": "./data/node1",
  "listen_addr": "127.0.0.1:7001",
  "http_addr": "127.0.0.1:8001",
  "peers": ["127.0.0.1:7002", "127.0.0.1:7003"],
  "genesis_file": "./configs/genesis.example.json",
  "block_interval_seconds": 8,
  "max_tx_per_block": 1024,
  "private_key_hex": "<validator-private-key-hex>",
  "validator_public_keys": ["<validator-1-public-key-hex>", "<validator-2-public-key-hex>", "<validator-3-public-key-hex>"]
}
```

Fields:

- `node_id`
  - unique node name used in logs and peer identity messages
  - required
- `data_dir`
  - local node data directory, including the embedded database
  - required
- `listen_addr`
  - TCP address for the P2P server
  - required
- `http_addr`
  - HTTP API listen address
  - optional; leave empty to disable the HTTP API
- `peers`
  - static peer addresses to connect to on startup
  - optional, but usually set for multi-node demos
- `genesis_file`
  - path to the genesis JSON file
  - required
  - relative paths are resolved relative to the node config file location
- `block_interval_seconds`
  - target block production interval in seconds
  - optional; defaults to `8`
- `max_tx_per_block`
  - maximum number of non-coinbase transactions selected from the mempool for a block
  - optional; defaults to `1024`
- `private_key_hex`
  - hex-encoded ed25519 private key for the local validator
  - required
  - must match one of the configured validator public keys
- `validator_public_keys`
  - list of hex-encoded ed25519 public keys for the fixed validator set
  - required
  - at least one validator key must be present

Notes:

- the current node implementation is validator-only
- peer discovery is static in the current design
- the node stores data in `data_dir` using bbolt

### Genesis Configuration

A genesis config defines the initial chain state.

Example shape:

```json
{
  "chain_id": "blockgo-local",
  "created_at_utc": "2026-01-01T00:00:00Z",
  "block_interval_seconds": 8,
  "validators": [
    { "id": "node1", "public_key": "<validator-1-public-key-hex>" },
    { "id": "node2", "public_key": "<validator-2-public-key-hex>" },
    { "id": "node3", "public_key": "<validator-3-public-key-hex>" }
  ],
  "allocations": [
    { "address": "<funded-address-hex>", "value": 1000 }
  ]
}
```

Fields:

- `chain_id`
  - human-readable chain identifier
  - required by the config format
- `created_at_utc`
  - informational UTC timestamp string in the JSON file
- `block_interval_seconds`
  - informational block interval value in the JSON file
- `validators`
  - initial fixed validator set
  - required
  - each validator contains:
    - `id`: human-readable validator name
    - `public_key`: hex-encoded ed25519 public key
- `allocations`
  - initial funded addresses in the genesis state
  - optional but usually present for demos and testing
  - each allocation contains:
    - `address`: destination address
    - `value`: initial balance

Important implementation detail:

The current genesis loader uses these fields to build the actual genesis block:

- `chain_id`
- `validators[*].public_key`
- `allocations[*]`

The following JSON fields are currently informational and are not enforced by genesis block construction:

- `created_at_utc`
- `block_interval_seconds`
- `validators[*].id`

They are still useful for readability and documentation, and they are included in the example and generated files.

### Validation Rules

At startup, BlockGo validates configuration strictly.

Examples of invalid configuration include:

- missing `node_id`
- missing `data_dir`
- missing `listen_addr`
- missing `genesis_file`
- invalid private key length
- empty validator set
- invalid validator public key length
- malformed hex values
- invalid allocation address format

## HTTP API Contract

The node exposes a small optional HTTP API.

Current endpoints:

- `GET /healthz`
- `GET /v1/chain/head`
- `GET /v1/mempool`
- `POST /v1/transactions`

### `GET /healthz`

Response:

```json
{
  "status": "ok"
}
```

### `GET /v1/chain/head`

Returns the current head block as JSON.

### `GET /v1/mempool`

Response shape:

```json
{
  "size": 0
}
```

### `POST /v1/transactions`

Request shape:

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

Accepted response:

```json
{
  "tx_id": "<tx-id-hex>"
}
```

Error response shape:

```json
{
  "error": "<message>"
}
```

Request handling rules:

- the request body must contain exactly one JSON object
- unknown JSON fields are rejected
- oversized request bodies are rejected before transaction decoding
- valid JSON is decoded before transaction-level validation runs

## Local Demo Workflow

The intended Docker demo flow is:

1. generate Docker-oriented config files with `go run ./cmd/blockgo gen-localnet -mode docker -nodes 3 -out ./configs/local`
2. start three nodes with Docker Compose
3. inspect health endpoints
4. submit transactions and observe block production

For a single local node without Docker, use `make run-node` or generate local-mode config with `go run ./cmd/blockgo gen-localnet -mode local -nodes 1 -out ./configs/run-node` and start `configs/run-node/node1.json` directly.

This keeps the onboarding path short while still demonstrating both local and multi-node behavior.

## Known Limits

The current system intentionally keeps these limits visible:

- fixed validator set
- static peer configuration
- minimal HTTP API
- no smart contracts or VM
- no dynamic validator membership
- no hostile-network protections
- educational focus over production hardening
