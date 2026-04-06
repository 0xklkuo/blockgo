package store

import (
	"path/filepath"
	"testing"
	"time"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
)

func TestStoreSaveGenesisAndLoadHead(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "blockgo.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	validatorPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(validatorPub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	genesis, err := blockchain.NewGenesisBlock(blockchain.GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{validatorPub},
		Allocations: []blockchain.GenesisAllocation{
			{Address: addr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	utxos := blockchain.NewUTXOSet()
	if err := utxos.ApplyGenesisBlock(genesis); err != nil {
		t.Fatalf("ApplyGenesisBlock() error = %v", err)
	}

	if err := s.SaveGenesis(genesis, utxos); err != nil {
		t.Fatalf("SaveGenesis() error = %v", err)
	}

	head, err := s.LoadHead()
	if err != nil {
		t.Fatalf("LoadHead() error = %v", err)
	}

	if head == nil {
		t.Fatal("LoadHead() returned nil head")
	}

	if head.Hash != genesis.Hash {
		t.Fatalf("head hash = %s, want %s", head.Hash.String(), genesis.Hash.String())
	}
}

func TestStoreSaveBlockAndReloadUTXOSet(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "blockgo.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

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

	genesis, err := blockchain.NewGenesisBlock(blockchain.GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{validatorPub},
		Allocations: []blockchain.GenesisAllocation{
			{Address: senderAddr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	utxos := blockchain.NewUTXOSet()
	if err := utxos.ApplyGenesisBlock(genesis); err != nil {
		t.Fatalf("ApplyGenesisBlock() error = %v", err)
	}

	if err := s.SaveGenesis(genesis, utxos); err != nil {
		t.Fatalf("SaveGenesis() error = %v", err)
	}

	prevTx := genesis.Transactions[0]

	spendTx := blockchain.Transaction{
		Inputs: []blockchain.TxInput{
			{
				PrevOut: blockchain.OutPoint{
					TxID:  prevTx.ID,
					Index: 0,
				},
				PublicKey: senderPub,
			},
		},
		Outputs: []blockchain.TxOutput{
			{Value: 60, Address: receiverAddr},
			{Value: 30, Address: senderAddr},
		},
	}
	if err := spendTx.SignInput(0, senderPriv); err != nil {
		t.Fatalf("SignInput() error = %v", err)
	}

	coinbaseTx := blockchain.Transaction{
		Outputs: []blockchain.TxOutput{
			{Value: blockchain.BlockReward + 10, Address: validatorAddr},
		},
	}
	if err := coinbaseTx.Finalize(); err != nil {
		t.Fatalf("coinbase Finalize() error = %v", err)
	}

	block, err := blockchain.NewBlock(
		1,
		genesis.Hash,
		validatorPub,
		[]blockchain.Transaction{coinbaseTx, spendTx},
		time.Unix(blockchain.GenesisTimestampUTC+5, 0).UTC(),
	)
	if err != nil {
		t.Fatalf("NewBlock() error = %v", err)
	}

	if err := blockchain.ValidateBlock(block, genesis, utxos); err != nil {
		t.Fatalf("ValidateBlock() error = %v", err)
	}

	nextUTXOs := utxos.Clone()
	if err := nextUTXOs.ApplyBlock(block); err != nil {
		t.Fatalf("ApplyBlock() error = %v", err)
	}

	if err := s.SaveBlock(block, nextUTXOs); err != nil {
		t.Fatalf("SaveBlock() error = %v", err)
	}

	loadedHead, err := s.LoadHead()
	if err != nil {
		t.Fatalf("LoadHead() error = %v", err)
	}

	if loadedHead == nil {
		t.Fatal("LoadHead() returned nil")
	}

	if loadedHead.Hash != block.Hash {
		t.Fatalf("loaded head hash = %s, want %s", loadedHead.Hash.String(), block.Hash.String())
	}

	loadedUTXOs, err := s.LoadUTXOSet()
	if err != nil {
		t.Fatalf("LoadUTXOSet() error = %v", err)
	}

	if loadedUTXOs.Size() != nextUTXOs.Size() {
		t.Fatalf("loaded utxo size = %d, want %d", loadedUTXOs.Size(), nextUTXOs.Size())
	}
}

func TestStoreReopenAndLoadState(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "blockgo.db")

	validatorPub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := blockcrypto.AddressFromPublicKey(validatorPub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	genesis, err := blockchain.NewGenesisBlock(blockchain.GenesisConfig{
		ChainID:    "blockgo-local",
		Validators: [][]byte{validatorPub},
		Allocations: []blockchain.GenesisAllocation{
			{Address: addr, Value: 100},
		},
	})
	if err != nil {
		t.Fatalf("NewGenesisBlock() error = %v", err)
	}

	utxos := blockchain.NewUTXOSet()
	if err := utxos.ApplyGenesisBlock(genesis); err != nil {
		t.Fatalf("ApplyGenesisBlock() error = %v", err)
	}

	{
		s, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}

		if err := s.SaveGenesis(genesis, utxos); err != nil {
			_ = s.Close()
			t.Fatalf("SaveGenesis() error = %v", err)
		}

		if err := s.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}

	{
		s, err := Open(dbPath)
		if err != nil {
			t.Fatalf("reopen Open() error = %v", err)
		}
		defer func() {
			if err := s.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		}()

		head, err := s.LoadHead()
		if err != nil {
			t.Fatalf("LoadHead() error = %v", err)
		}

		if head == nil {
			t.Fatal("LoadHead() returned nil")
		}

		if head.Hash != genesis.Hash {
			t.Fatalf("head hash = %s, want %s", head.Hash.String(), genesis.Hash.String())
		}

		loadedUTXOs, err := s.LoadUTXOSet()
		if err != nil {
			t.Fatalf("LoadUTXOSet() error = %v", err)
		}

		if loadedUTXOs.Size() != 1 {
			t.Fatalf("utxo size = %d, want 1", loadedUTXOs.Size())
		}
	}
}
