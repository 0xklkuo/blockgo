package blockchain

import (
	"testing"

	blockcrypto "blockgo/internal/crypto"
)

func TestApplyGenesisBlock(t *testing.T) {
	pub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	genesis, err := NewGenesisBlock(GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{pub},
		Allocations: []GenesisAllocation{
			{Address: addr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	utxos := NewUTXOSet()
	if err := utxos.ApplyGenesisBlock(genesis); err != nil {
		t.Fatalf("ApplyGenesisBlock() error = %v", err)
	}

	if utxos.Size() != 1 {
		t.Fatalf("utxo size = %d, want 1", utxos.Size())
	}
}

func TestApplyBlockConsumesAndCreatesUTXOs(t *testing.T) {
	senderPub, senderPriv, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	receiverPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	validatorPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	senderAddr, _ := blockcrypto.AddressFromPublicKey(senderPub)
	receiverAddr, _ := blockcrypto.AddressFromPublicKey(receiverPub)
	validatorAddr, _ := blockcrypto.AddressFromPublicKey(validatorPub)

	genesis, err := NewGenesisBlock(GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{validatorPub},
		Allocations: []GenesisAllocation{
			{Address: senderAddr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	utxos := NewUTXOSet()
	if err := utxos.ApplyGenesisBlock(genesis); err != nil {
		t.Fatalf("ApplyGenesisBlock() error = %v", err)
	}

	prevTx := genesis.Transactions[0]

	spendTx := Transaction{
		Inputs: []TxInput{
			{
				PrevOut: OutPoint{
					TxID:  prevTx.ID,
					Index: 0,
				},
				PublicKey: senderPub,
			},
		},
		Outputs: []TxOutput{
			{Value: 60, Address: receiverAddr},
			{Value: 30, Address: senderAddr},
		},
	}

	if err := spendTx.SignInput(0, senderPriv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	coinbaseTx := Transaction{
		Outputs: []TxOutput{
			{Value: BlockReward + 10, Address: validatorAddr},
		},
	}
	if err := coinbaseTx.Finalize(); err != nil {
		t.Fatalf("coinbase Finalize() error = %v", err)
	}

	block, err := NewBlock(1, genesis.Hash, validatorPub, []Transaction{coinbaseTx, spendTx}, mustTime(GenesisTimestampUTC+5))
	if err != nil {
		t.Fatalf("NewBlock() error = %v", err)
	}

	if err := ValidateBlock(block, genesis, utxos); err != nil {
		t.Fatalf("ValidateBlock() error = %v", err)
	}

	if err := utxos.ApplyBlock(block); err != nil {
		t.Fatalf("ApplyBlock() error = %v", err)
	}

	if _, ok := utxos.Get(OutPoint{TxID: prevTx.ID, Index: 0}); ok {
		t.Fatal("spent genesis output still exists")
	}

	if utxos.Size() != 3 {
		t.Fatalf("utxo size = %d, want 3", utxos.Size())
	}
}
