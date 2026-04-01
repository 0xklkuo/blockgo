# BlockGo Architecture

## Purpose

BlockGo is a minimal educational blockchain in Go that aims to demonstrate real core concepts without excessive complexity.

## Design Priorities

1. Simplicity
2. Correctness
3. Readability
4. Contributor friendliness
5. Realistic structure
6. Minimal dependencies

## V1 System Shape

### Ledger Model
UTXO-based ledger.

Why:
- teaches transaction validation directly
- teaches spend rules and ownership proofs
- makes double-spend handling explicit

### Consensus
Proof of Authority with:
- fixed validator set
- deterministic round-robin proposer schedule
- signed blocks

Why:
- efficient
- easy to explain
- suitable for multi-node local demos
- much simpler than full Proof of Stake

### Networking
Custom TCP-based P2P with:
- static peers
- small message set
- readable protocol

Why not libp2p in v1:
- too much abstraction for a teaching-first minimal project

### Storage
bbolt embedded key-value store.

Expected persisted data:
- blocks
- metadata
- UTXO set
- transaction index

### Crypto
- sha256 for hashes
- ed25519 for signatures

## Package Direction

```text
internal/blockchain   core types and validation
internal/consensus    validator/proposer rules
internal/crypto       keys, addresses, signatures
internal/mempool      pending transactions
internal/node         node lifecycle
internal/p2p          peer protocol and sync
internal/store        persistence
internal/config       config parsing/validation
```

## Non-Goals for V1

- Proof of Work
- full Proof of Stake economics
- Merkle Patricia Trie
- smart contracts
- dynamic validator changes
- advanced fork choice / finality protocols
- production-grade discovery and anti-abuse features

## API Philosophy

Keep boundaries explicit:
- blockchain logic should not depend on HTTP
- storage should not own consensus logic
- p2p should transport validated domain messages
- config should be simple JSON

## Expected Growth Path

- M1: primitives
- M2: UTXO validation
- M3: persistence
- M4: node loop and PoA
- M5: peer sync
- M6: CLI and Docker
- M7: polish
