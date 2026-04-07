package p2p_test

import (
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"blockgo/internal/blockchain"
	"blockgo/internal/config"
	blockcrypto "blockgo/internal/crypto"
	"blockgo/internal/node"
)

func TestTwoNodesSyncBlock(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	v1Pub, v1Priv, _ := blockcrypto.GenerateKeyPair()
	v2Pub, v2Priv, _ := blockcrypto.GenerateKeyPair()
	receiverPub, _, _ := blockcrypto.GenerateKeyPair()

	v1Addr, _ := blockcrypto.AddressFromPublicKey(v1Pub)
	receiverAddr, _ := blockcrypto.AddressFromPublicKey(receiverPub)

	genesis, err := blockchain.NewGenesisBlock(blockchain.GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{v1Pub, v2Pub},
		Allocations: []blockchain.GenesisAllocation{
			{Address: v1Addr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	node1, err := node.New(config.NodeConfig{
		NodeID:              "node1",
		DataDir:             filepath.Join(t.TempDir(), "node1"),
		ListenAddr:          "127.0.0.1:9101",
		Peers:               []string{"127.0.0.1:9102"},
		BlockInterval:       time.Hour,
		MaxTxPerBlock:       100,
		LocalPrivateKey:     v1Priv,
		ValidatorPublicKeys: [][]byte{v1Pub, v2Pub},
	}, genesis, logger)
	if err != nil {
		t.Fatalf("node.New(node1) error = %v", err)
	}
	defer func() { _ = node1.Stop() }()

	node2, err := node.New(config.NodeConfig{
		NodeID:              "node2",
		DataDir:             filepath.Join(t.TempDir(), "node2"),
		ListenAddr:          "127.0.0.1:9102",
		Peers:               []string{"127.0.0.1:9101"},
		BlockInterval:       time.Hour,
		MaxTxPerBlock:       100,
		LocalPrivateKey:     v2Priv,
		ValidatorPublicKeys: [][]byte{v1Pub, v2Pub},
	}, genesis, logger)
	if err != nil {
		t.Fatalf("node.New(node2) error = %v", err)
	}
	defer func() { _ = node2.Stop() }()

	if err := node1.Start(); err != nil {
		t.Fatalf("node1.Start() error = %v", err)
	}
	if err := node2.Start(); err != nil {
		t.Fatalf("node2.Start() error = %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	genesisTx := genesis.Transactions[0]
	tx := blockchain.Transaction{
		Inputs: []blockchain.TxInput{
			{
				PrevOut: blockchain.OutPoint{
					TxID:  genesisTx.ID,
					Index: 0,
				},
				PublicKey: v1Pub,
			},
		},
		Outputs: []blockchain.TxOutput{
			{Value: 60, Address: receiverAddr},
			{Value: 30, Address: v1Addr},
		},
	}
	if err := tx.SignInput(0, v1Priv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	if err := node1.SubmitTransaction(tx); err != nil {
		t.Fatalf("node1.SubmitTransaction() error = %v", err)
	}

	block, err := node1.TryProduceBlock(time.Unix(blockchain.GenesisTimestampUTC+5, 0).UTC())
	if err != nil {
		t.Fatalf("node1.TryProduceBlock() error = %v", err)
	}
	if block == nil {
		t.Fatal("node1.TryProduceBlock() returned nil block")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if node2.Head().Header.Height == 1 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("node2 head height = %d, want 1", node2.Head().Header.Height)
}
