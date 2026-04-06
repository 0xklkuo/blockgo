package consensus

import (
	"testing"
	"time"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
)

func TestProposerRotation(t *testing.T) {
	v1, _, _ := blockcrypto.GenerateKeyPair()
	v2, _, _ := blockcrypto.GenerateKeyPair()
	v3, _, _ := blockcrypto.GenerateKeyPair()

	set, err := NewValidatorSet([][]byte{v1, v2, v3})
	if err != nil {
		t.Fatalf("NewValidatorSet() error = %v", err)
	}

	tests := []struct {
		height uint64
		want   []byte
	}{
		{1, v1},
		{2, v2},
		{3, v3},
		{4, v1},
		{5, v2},
	}

	for _, tt := range tests {
		got, err := set.Proposer(tt.height)
		if err != nil {
			t.Fatalf("Proposer(%d) error = %v", tt.height, err)
		}
		if string(got) != string(tt.want) {
			t.Fatalf("Proposer(%d) mismatch", tt.height)
		}
	}
}

func TestSignAndValidateBlock(t *testing.T) {
	v1Pub, v1Priv, _ := blockcrypto.GenerateKeyPair()
	v2Pub, _, _ := blockcrypto.GenerateKeyPair()

	set, err := NewValidatorSet([][]byte{v1Pub, v2Pub})
	if err != nil {
		t.Fatalf("NewValidatorSet() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(v1Pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	genesis, err := blockchain.NewGenesisBlock(blockchain.GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{v1Pub, v2Pub},
		Allocations: []blockchain.GenesisAllocation{
			{Address: addr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	block, err := blockchain.NewBlock(
		1,
		genesis.Hash,
		v1Pub,
		[]blockchain.Transaction{
			{Outputs: []blockchain.TxOutput{{Value: blockchain.BlockReward, Address: addr}}},
		},
		time.Unix(blockchain.GenesisTimestampUTC+5, 0).UTC(),
	)
	if err != nil {
		t.Fatalf("NewBlock() error = %v", err)
	}

	if err := set.SignBlock(block, v1Priv); err != nil {
		t.Fatalf("SignBlock() error = %v", err)
	}

	if err := set.ValidateBlock(block); err != nil {
		t.Fatalf("ValidateBlock() error = %v", err)
	}
}
