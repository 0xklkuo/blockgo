package blockchain

import (
	"testing"
	"time"

	blockcrypto "blockgo/internal/crypto"
)

func TestValidateTransactionAcceptsValidSpend(t *testing.T) {
	senderPub, senderPriv, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	receiverPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	senderAddr, _ := blockcrypto.AddressFromPublicKey(senderPub)
	receiverAddr, _ := blockcrypto.AddressFromPublicKey(receiverPub)

	utxos := NewUTXOSet()

	prevTx := Transaction{
		Outputs: []TxOutput{
			{Value: 100, Address: senderAddr},
		},
	}
	if err := prevTx.Finalize(); err != nil {
		t.Fatalf("prevTx.Finalize() error = %v", err)
	}

	utxos.Put(OutPoint{TxID: prevTx.ID, Index: 0}, prevTx.Outputs[0])

	tx := Transaction{
		Inputs: []TxInput{
			{
				PrevOut:   OutPoint{TxID: prevTx.ID, Index: 0},
				PublicKey: senderPub,
			},
		},
		Outputs: []TxOutput{
			{Value: 70, Address: receiverAddr},
			{Value: 20, Address: senderAddr},
		},
	}

	if err := tx.SignInput(0, senderPriv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	fee, err := ValidateTransaction(tx, utxos, map[UTXOKey]struct{}{})
	if err != nil {
		t.Fatalf("ValidateTransaction() error = %v", err)
	}

	if fee != 10 {
		t.Fatalf("fee = %d, want 10", fee)
	}
}

func TestValidateTransactionRejectsWrongSigner(t *testing.T) {
	senderPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	wrongPub, wrongPriv, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	receiverPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	senderAddr, _ := blockcrypto.AddressFromPublicKey(senderPub)
	receiverAddr, _ := blockcrypto.AddressFromPublicKey(receiverPub)

	utxos := NewUTXOSet()

	prevTx := Transaction{
		Outputs: []TxOutput{
			{Value: 100, Address: senderAddr},
		},
	}
	if err := prevTx.Finalize(); err != nil {
		t.Fatalf("prevTx.Finalize() error = %v", err)
	}

	utxos.Put(OutPoint{TxID: prevTx.ID, Index: 0}, prevTx.Outputs[0])

	tx := Transaction{
		Inputs: []TxInput{
			{
				PrevOut:   OutPoint{TxID: prevTx.ID, Index: 0},
				PublicKey: wrongPub,
			},
		},
		Outputs: []TxOutput{
			{Value: 90, Address: receiverAddr},
		},
	}

	if err := tx.SignInput(0, wrongPriv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	if _, err := ValidateTransaction(tx, utxos, map[UTXOKey]struct{}{}); err == nil {
		t.Fatal("ValidateTransaction() error = nil, want non-nil")
	}
}

func TestValidateTransactionRejectsOverspend(t *testing.T) {
	senderPub, senderPriv, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	receiverPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	senderAddr, _ := blockcrypto.AddressFromPublicKey(senderPub)
	receiverAddr, _ := blockcrypto.AddressFromPublicKey(receiverPub)

	utxos := NewUTXOSet()

	prevTx := Transaction{
		Outputs: []TxOutput{
			{Value: 100, Address: senderAddr},
		},
	}
	if err := prevTx.Finalize(); err != nil {
		t.Fatalf("prevTx.Finalize() error = %v", err)
	}

	utxos.Put(OutPoint{TxID: prevTx.ID, Index: 0}, prevTx.Outputs[0])

	tx := Transaction{
		Inputs: []TxInput{
			{
				PrevOut:   OutPoint{TxID: prevTx.ID, Index: 0},
				PublicKey: senderPub,
			},
		},
		Outputs: []TxOutput{
			{Value: 101, Address: receiverAddr},
		},
	}

	if err := tx.SignInput(0, senderPriv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	if _, err := ValidateTransaction(tx, utxos, map[UTXOKey]struct{}{}); err == nil {
		t.Fatal("ValidateTransaction() error = nil, want non-nil")
	}
}

func TestValidateTransactionRejectsDoubleSpendWithinBlock(t *testing.T) {
	senderPub, senderPriv, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	receiverPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	senderAddr, _ := blockcrypto.AddressFromPublicKey(senderPub)
	receiverAddr, _ := blockcrypto.AddressFromPublicKey(receiverPub)

	utxos := NewUTXOSet()

	prevTx := Transaction{
		Outputs: []TxOutput{
			{Value: 100, Address: senderAddr},
		},
	}
	if err := prevTx.Finalize(); err != nil {
		t.Fatalf("prevTx.Finalize() error = %v", err)
	}

	outPoint := OutPoint{TxID: prevTx.ID, Index: 0}
	utxos.Put(outPoint, prevTx.Outputs[0])

	tx1 := Transaction{
		Inputs: []TxInput{
			{
				PrevOut:   outPoint,
				PublicKey: senderPub,
			},
		},
		Outputs: []TxOutput{
			{Value: 90, Address: receiverAddr},
		},
	}
	if err := tx1.SignInput(0, senderPriv); err != nil {
		t.Fatalf("tx1.SignInput() error = %v", err)
	}

	spent := map[UTXOKey]struct{}{}
	if _, err := ValidateTransaction(tx1, utxos, spent); err != nil {
		t.Fatalf("ValidateTransaction(tx1) error = %v", err)
	}

	tx2 := Transaction{
		Inputs: []TxInput{
			{
				PrevOut:   outPoint,
				PublicKey: senderPub,
			},
		},
		Outputs: []TxOutput{
			{Value: 80, Address: receiverAddr},
		},
	}
	if err := tx2.SignInput(0, senderPriv); err != nil {
		t.Fatalf("tx2.SignInput() error = %v", err)
	}

	if _, err := ValidateTransaction(tx2, utxos, spent); err == nil {
		t.Fatal("ValidateTransaction(tx2) error = nil, want non-nil")
	}
}

func TestValidateCoinbaseTransaction(t *testing.T) {
	pub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, _ := blockcrypto.AddressFromPublicKey(pub)

	tx := Transaction{
		Outputs: []TxOutput{
			{Value: BlockReward + 7, Address: addr},
		},
	}
	if err := tx.Finalize(); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}

	if err := ValidateCoinbaseTransaction(tx, 7); err != nil {
		t.Fatalf("ValidateCoinbaseTransaction() error = %v", err)
	}
}

func TestValidateBlockAcceptsValidBlock(t *testing.T) {
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
				PrevOut:   OutPoint{TxID: prevTx.ID, Index: 0},
				PublicKey: senderPub,
			},
		},
		Outputs: []TxOutput{
			{Value: 70, Address: receiverAddr},
			{Value: 20, Address: senderAddr},
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

	block, err := NewBlock(
		1,
		genesis.Hash,
		validatorPub,
		[]Transaction{coinbaseTx, spendTx},
		time.Unix(GenesisTimestampUTC+5, 0).UTC(),
	)
	if err != nil {
		t.Fatalf("NewBlock() error = %v", err)
	}

	if err := ValidateBlock(block, genesis, utxos); err != nil {
		t.Fatalf("ValidateBlock() error = %v", err)
	}
}
