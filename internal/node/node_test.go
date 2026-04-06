package node

import (
	"testing"
	"time"

	"blockgo/internal/blockchain"
	"blockgo/internal/config"
	blockcrypto "blockgo/internal/crypto"
)

func TestTryProduceBlock(t *testing.T) {
	validatorPub, validatorPriv, _ := blockcrypto.GenerateKeyPair()
	receiverPub, _, _ := blockcrypto.GenerateKeyPair()

	validatorAddr, _ := blockcrypto.AddressFromPublicKey(validatorPub)
	receiverAddr, _ := blockcrypto.AddressFromPublicKey(receiverPub)

	genesis, err := blockchain.NewGenesisBlock(blockchain.GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{validatorPub},
		Allocations: []blockchain.GenesisAllocation{
			{Address: validatorAddr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	cfg := config.NodeConfig{
		NodeID:              "validator-1",
		BlockInterval:       1 * time.Second,
		MaxTxPerBlock:       100,
		LocalPrivateKey:     validatorPriv,
		ValidatorPublicKeys: [][]byte{validatorPub},
	}

	n, err := New(cfg, genesis)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Spend genesis allocation with fee 10
	genesisTx := genesis.Transactions[0]
	tx := blockchain.Transaction{
		Inputs: []blockchain.TxInput{
			{
				PrevOut: blockchain.OutPoint{
					TxID:  genesisTx.ID,
					Index: 0,
				},
				PublicKey: validatorPub,
			},
		},
		Outputs: []blockchain.TxOutput{
			{Value: 60, Address: receiverAddr},
			{Value: 30, Address: validatorAddr},
		},
	}
	if err := tx.SignInput(0, validatorPriv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf("SubmitTransaction() error = %v", err)
	}

	block, err := n.TryProduceBlock(time.Unix(blockchain.GenesisTimestampUTC+5, 0).UTC())
	if err != nil {
		t.Fatalf("TryProduceBlock() error = %v", err)
	}

	if block == nil {
		t.Fatal("TryProduceBlock() returned nil block")
	}

	if block.Header.Height != 1 {
		t.Fatalf("block height = %d, want 1", block.Header.Height)
	}

	if len(block.Transactions) != 2 {
		t.Fatalf("block tx count = %d, want 2 (coinbase + spend)", len(block.Transactions))
	}

	if n.MempoolLen() != 0 {
		t.Fatalf("mempool len = %d, want 0", n.MempoolLen())
	}
}
