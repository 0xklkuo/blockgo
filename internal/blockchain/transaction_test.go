package blockchain

import (
	"testing"

	blockcrypto "blockgo/internal/crypto"
)

func TestTransactionFinalizeDeterministic(t *testing.T) {
	pub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	tx1 := Transaction{
		Inputs: nil,
		Outputs: []TxOutput{
			{Value: 100, Address: addr},
		},
	}

	tx2 := Transaction{
		Inputs: nil,
		Outputs: []TxOutput{
			{Value: 100, Address: addr},
		},
	}

	if err := tx1.Finalize(); err != nil {
		t.Fatalf("tx1.Finalize() error = %v", err)
	}

	if err := tx2.Finalize(); err != nil {
		t.Fatalf("tx2.Finalize() error = %v", err)
	}

	if tx1.ID != tx2.ID {
		t.Fatalf("tx ids differ: %s != %s", tx1.ID.String(), tx2.ID.String())
	}
}

func TestTransactionSignAndVerifyInput(t *testing.T) {
	pub, priv, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	tx := Transaction{
		Inputs: []TxInput{
			{
				PrevOut: OutPoint{
					TxID:  Hash{1, 2, 3},
					Index: 0,
				},
				PublicKey: pub,
			},
		},
		Outputs: []TxOutput{
			{Value: 50, Address: addr},
		},
	}

	if err := tx.SignInput(0, priv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	if !tx.VerifyInputSignature(0) {
		t.Fatal("VerifyInputSignature() = false, want true")
	}
}

func TestCoinbaseDetection(t *testing.T) {
	tx := Transaction{
		Inputs: nil,
		Outputs: []TxOutput{
			{Value: 1},
		},
	}

	if !tx.IsCoinbase() {
		t.Fatal("IsCoinbase() = false, want true")
	}
}
