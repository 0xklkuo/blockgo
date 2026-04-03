package blockchain

import (
	"testing"

	blockcrypto "blockgo/internal/crypto"
)

func TestMerkleRootEmpty(t *testing.T) {
	root := MerkleRoot(nil)
	if IsZeroHash(root) {
		t.Fatal("empty merkle root should not be zero hash")
	}
}

func TestMerkleRootDeterministic(t *testing.T) {
	pub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	tx1 := Transaction{
		Outputs: []TxOutput{{Value: 10, Address: addr}},
	}
	tx2 := Transaction{
		Outputs: []TxOutput{{Value: 20, Address: addr}},
	}

	if err := tx1.Finalize(); err != nil {
		t.Fatalf("tx1.Finalize() error = %v", err)
	}
	if err := tx2.Finalize(); err != nil {
		t.Fatalf("tx2.Finalize() error = %v", err)
	}

	root1 := MerkleRoot([]Transaction{tx1, tx2})
	root2 := MerkleRoot([]Transaction{tx1, tx2})

	if root1 != root2 {
		t.Fatalf("roots differ: %s != %s", root1.String(), root2.String())
	}
}
