# Configuration

BlockGo uses JSON configuration files for both node startup and genesis state.

This document describes the current schema used by the codebase and the example files in `configs/`.

## Files

- `configs/node.example.json`
- `configs/genesis.example.json`

For a runnable local multi-node demo, generate config files with:

```bash
go run ./cmd/blockgo gen-localnet -out ./configs/local
```

That command creates:

- `configs/local/genesis.json`
- `configs/local/node1.json`
- `configs/local/node2.json`
- `configs/local/node3.json`

These generated files are for local development only and include demo private keys.

## Node Configuration

A node config file defines how a single node starts.

### Schema

```bash
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

### Fields

- `node_id`
  - Unique node name used in logs and peer identity messages.
  - Required.

- `data_dir`
  - Directory used for local node data, including the embedded database.
  - Required.

- `listen_addr`
  - TCP address for the P2P server.
  - Required.

- `http_addr`
  - HTTP API listen address.
  - Optional. Leave empty to disable the HTTP API.

- `peers`
  - Static peer addresses to connect to on startup.
  - Optional, but usually set for multi-node demos.

- `genesis_file`
  - Path to the genesis JSON file.
  - Required.
  - Relative paths are resolved relative to the node config file location.

- `block_interval_seconds`
  - Target block production interval in seconds.
  - Optional.
  - If `0` or omitted, the default is `8`.

- `max_tx_per_block`
  - Maximum number of non-coinbase transactions selected from the mempool for a block.
  - Optional.
  - If `0` or omitted, the default is `1024`.

- `private_key_hex`
  - Hex-encoded ed25519 private key for the local validator.
  - Required.
  - Must match one of the configured validator public keys.

- `validator_public_keys`
  - List of hex-encoded ed25519 public keys for the fixed validator set.
  - Required.
  - At least one validator key must be present.

### Notes

- The current node implementation is validator-only. A node must start with a validator private key that belongs to the configured validator set.
- Peer discovery is static in v1. Nodes connect only to the peers listed in `peers`.
- The node stores data in `data_dir` using bbolt.

## Genesis Configuration

A genesis config file defines the initial chain state.

### Schema

```bash
{
  "chain_id": "blockgo-local",
  "created_at_utc": "2026-01-01T00:00:00Z",
  "block_interval_seconds": 8,
  "validators": [
    { "id": "node1", "public_key": "<validator-1-public-key-hex>" },
    { "id": "node2", "public_key": "<validator-2-public-key-hex>" },
    { "id": "node3", "public_key": "<validator-3-public-key-hex>" }
  ],
  "allocations": [{ "address": "<funded-address-hex>", "value": 1000 }]
}
```

### Fields

- `chain_id`
  - Human-readable chain identifier.
  - Required by the config format.

- `created_at_utc`
  - Informational UTC timestamp string in the JSON file.
  - Present in examples and generated local demo configs.

- `block_interval_seconds`
  - Informational block interval value in the JSON file.
  - Present in examples and generated local demo configs.

- `validators`
  - Initial fixed validator set.
  - Required.
  - Each validator entry contains:
    - `id`: human-readable validator name
    - `public_key`: hex-encoded ed25519 public key

- `allocations`
  - Initial funded addresses in the genesis state.
  - Optional but usually present for demos and testing.
  - Each allocation contains:
    - `address`: destination address
    - `value`: initial balance

## Important Implementation Detail

The current genesis loader uses these fields to build the actual genesis block:

- `chain_id`
- `validators[*].public_key`
- `allocations[*]`

The following JSON fields are currently informational and are not enforced by genesis block construction:

- `created_at_utc`
- `block_interval_seconds`
- `validators[*].id`

They are still useful for readability and documentation, and they are included in the example and generated files.

## Validation Rules

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

## Example Workflow

### Use tracked example files

1. Copy the example files.
2. Replace placeholder keys and addresses.
3. Start a node with:

```bash
go run ./cmd/blockgo-node -config ./configs/node.example.json
```

### Use generated local demo files

1. Generate local config:

```bash
go run ./cmd/blockgo gen-localnet -out ./configs/local
```

2. Start the demo network:

```bash
docker compose up --build
```

3. Check health:

```bash
curl http://localhost:8001/healthz
curl http://localhost:8002/healthz
curl http://localhost:8003/healthz
```

## Design Notes

The config format is intentionally small and explicit:

- JSON only
- fixed validator set
- static peers
- simple startup validation
- human-readable example files

This keeps the project easy to understand, easy to edit, and suitable for educational use.
