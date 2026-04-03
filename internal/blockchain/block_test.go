package blockchain

import (
	"testing"
	"time"

	blockcrypto "blockgo/internal/crypto"
)

func TestNewBlockFinalizesHashAndMerkleRoot(t *testing.T) {
	pub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	tx := Transaction{
		Outputs: []TxOutput{
			{Value: 100, Address: addr},
		},
	}

	block, err := NewBlock(1, Hash{9, 9, 9}, pub, []Transaction{tx}, time.Unix(100, 0))
	if err != nil {
		t.Fatalf("NewBlock() error = %v", err)
	}

	if IsZeroHash(block.Hash) {
		t.Fatal("block hash is zero")
	}

	if IsZeroHash(block.Header.MerkleRoot) {
		t.Fatal("merkle root is zero")
	}

	if len(block.Transactions) != 1 {
		t.Fatalf("len(block.Transactions) = %d, want 1", len(block.Transactions))
	}

	if IsZeroHash(block.Transactions[0].ID) {
		t.Fatal("transaction id is zero")
	}
}
