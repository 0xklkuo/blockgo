# Configuration Format

## Decision

blockgo uses **JSON** for configuration in v1.

## Why JSON

- supported directly by Go standard library
- easy to read and edit
- good enough for a small educational project
- avoids extra dependency or parser complexity

## Planned Config Files

### Node Config
Contains:
- node id
- data directory
- listen address
- peer list
- validator settings
- genesis file path
- logging level
- optional HTTP address

### Genesis Config
Contains:
- chain id
- validator set
- initial allocations / genesis UTXOs
- block interval

## Example Evolution

The exact schema may grow slightly as implementation progresses, but these principles should remain stable:
- keep names explicit
- avoid nested complexity
- keep values human-readable
- validate strictly at startup
