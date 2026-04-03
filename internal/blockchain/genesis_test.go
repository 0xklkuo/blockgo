package blockchain

import (
	"testing"

	blockcrypto "blockgo/internal/crypto"
)

func TestNewGenesisBlock(t *testing.T) {
	pub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	block, err := NewGenesisBlock(GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{pub},
		Allocations: []GenesisAllocation{
			{
				Address: addr,
				Value:   1000,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	if block.Header.Height != 0 {
		t.Fatalf("genesis height = %d, want 0", block.Header.Height)
	}

	if len(block.Transactions) != 1 {
		t.Fatalf("len(block.Transactions) = %d, want 1", len(block.Transactions))
	}

	if !block.Transactions[0].IsCoinbase() {
		t.Fatal("genesis transaction should be coinbase-like")
	}
}
