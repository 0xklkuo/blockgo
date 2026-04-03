package blockchain

import (
	"errors"
	"time"

	blockcrypto "blockgo/internal/crypto"
)

const GenesisTimestampUTC int64 = 1767225600 // 2026-01-01T00:00:00Z

type GenesisAllocation struct {
	Address blockcrypto.Address
	Value   uint64
}

type GenesisConfig struct {
	ChainID     string
	Validators  [][]byte
	Allocations []GenesisAllocation
}

func NewGenesisBlock(cfg GenesisConfig) (*Block, error) {
	if cfg.ChainID == "" {
		return nil, errors.New("chain id is required")
	}

	if len(cfg.Validators) == 0 {
		return nil, errors.New("at least one validator is required")
	}

	outputs := make([]TxOutput, 0, len(cfg.Allocations))
	for _, alloc := range cfg.Allocations {
		outputs = append(outputs, TxOutput{
			Value:   alloc.Value,
			Address: alloc.Address,
		})
	}

	genesisTx := Transaction{
		Inputs:  nil,
		Outputs: outputs,
	}

	if err := genesisTx.Finalize(); err != nil {
		return nil, err
	}

	return NewBlock(
		0,
		Hash{},
		cfg.Validators[0],
		[]Transaction{genesisTx},
		time.Unix(GenesisTimestampUTC, 0).UTC(),
	)
}
