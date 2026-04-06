package mempool

import (
	"testing"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
)

func TestAddAndRemove(t *testing.T) {
	pub, _, _ := blockcrypto.GenerateKeyPair()
	addr, _ := blockcrypto.AddressFromPublicKey(pub)

	tx := blockchain.Transaction{
		Outputs: []blockchain.TxOutput{{Value: 1, Address: addr}},
	}
	if err := tx.Finalize(); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}

	// Make it non-coinbase by adding a dummy input
	tx.Inputs = []blockchain.TxInput{
		{
			PrevOut:   blockchain.OutPoint{TxID: blockchain.Hash{1}, Index: 0},
			PublicKey: pub,
		},
	}
	if err := tx.Finalize(); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}

	p := New()
	if err := p.Add(tx); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if got := p.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}

	p.Remove(tx.ID)

	if got := p.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}
