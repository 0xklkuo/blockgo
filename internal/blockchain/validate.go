package blockchain

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	blockcrypto "blockgo/internal/crypto"
)

const BlockReward uint64 = 50

func ValidateGenesisBlock(block *Block) error {
	if block == nil {
		return errors.New("nil block")
	}

	if block.Header.Height != 0 {
		return errors.New("genesis block height must be 0")
	}

	if len(block.Transactions) != 1 {
		return errors.New("genesis block must contain exactly one transaction")
	}

	if block.Header.TimestampUTC != GenesisTimestampUTC {
		return errors.New("unexpected genesis timestamp")
	}

	tx := block.Transactions[0]
	if !tx.IsCoinbase() {
		return errors.New("genesis transaction must be coinbase-like")
	}

	if len(tx.Outputs) == 0 {
		return errors.New("genesis transaction must have outputs")
	}

	expectedMerkle := MerkleRoot(block.Transactions)
	if block.Header.MerkleRoot != expectedMerkle {
		return errors.New("invalid genesis merkle root")
	}

	expectedHash, err := block.ComputeHash()
	if err != nil {
		return fmt.Errorf("compute genesis hash: %w", err)
	}

	if block.Hash != expectedHash {
		return errors.New("invalid genesis block hash")
	}

	return nil
}

func ValidateBlock(block *Block, prevBlock *Block, utxos *UTXOSet) error {
	if block == nil {
		return errors.New("nil block")
	}

	if utxos == nil {
		return errors.New("nil utxo set")
	}

	if block.Header.Height == 0 {
		return ValidateGenesisBlock(block)
	}

	if prevBlock == nil {
		return errors.New("previous block is required for non-genesis block")
	}

	if block.Header.Height != prevBlock.Header.Height+1 {
		return errors.New("invalid block height")
	}

	if block.Header.PrevBlockHash != prevBlock.Hash {
		return errors.New("previous block hash mismatch")
	}

	if block.Header.TimestampUTC < prevBlock.Header.TimestampUTC {
		return errors.New("block timestamp moved backwards")
	}

	if len(block.Transactions) == 0 {
		return errors.New("block must contain at least one transaction")
	}

	expectedMerkle := MerkleRoot(block.Transactions)
	if block.Header.MerkleRoot != expectedMerkle {
		return errors.New("invalid merkle root")
	}

	expectedHash, err := block.ComputeHash()
	if err != nil {
		return fmt.Errorf("compute block hash: %w", err)
	}

	if block.Hash != expectedHash {
		return errors.New("invalid block hash")
	}

	if !block.Transactions[0].IsCoinbase() {
		return errors.New("first transaction must be coinbase")
	}

	for i := 1; i < len(block.Transactions); i++ {
		if block.Transactions[i].IsCoinbase() {
			return errors.New("multiple coinbase transactions are not allowed")
		}
	}

	working := utxos.Clone()
	totalFees := uint64(0)
	spentInBlock := make(map[UTXOKey]struct{})

	for i, tx := range block.Transactions {
		if i == 0 {
			continue
		}

		fee, err := ValidateTransaction(tx, working, spentInBlock)
		if err != nil {
			return fmt.Errorf("validate tx %d: %w", i, err)
		}

		totalFees += fee

		if err := ApplyTransaction(working, tx); err != nil {
			return fmt.Errorf("apply tx %d: %w", i, err)
		}
	}

	if err := ValidateCoinbaseTransaction(block.Transactions[0], totalFees); err != nil {
		return fmt.Errorf("validate coinbase: %w", err)
	}

	if err := ApplyTransaction(working, block.Transactions[0]); err != nil {
		return fmt.Errorf("apply coinbase: %w", err)
	}

	return nil
}

func ValidateTransaction(tx Transaction, utxos *UTXOSet, spentInBlock map[UTXOKey]struct{}) (uint64, error) {
	if utxos == nil {
		return 0, errors.New("nil utxo set")
	}

	if tx.IsCoinbase() {
		return 0, errors.New("coinbase transaction is not valid in regular tx validation")
	}

	if len(tx.Inputs) == 0 {
		return 0, errors.New("transaction must have at least one input")
	}

	if len(tx.Outputs) == 0 {
		return 0, errors.New("transaction must have at least one output")
	}

	expectedID, err := tx.ComputeID()
	if err != nil {
		return 0, fmt.Errorf("compute tx id: %w", err)
	}

	if tx.ID != expectedID {
		return 0, errors.New("invalid transaction id")
	}

	var inputSum uint64
	var outputSum uint64

	for _, out := range tx.Outputs {
		outputSum += out.Value
	}

	for inputIndex, in := range tx.Inputs {
		key := UTXOKey{
			TxID:  in.PrevOut.TxID,
			Index: in.PrevOut.Index,
		}

		if _, exists := spentInBlock[key]; exists {
			return 0, errors.New("double spend detected within block")
		}

		referencedOutput, ok := utxos.Get(in.PrevOut)
		if !ok {
			return 0, fmt.Errorf("referenced utxo not found: %s:%d", in.PrevOut.TxID.String(), in.PrevOut.Index)
		}

		addr, err := blockcrypto.AddressFromPublicKey(ed25519.PublicKey(in.PublicKey))
		if err != nil {
			return 0, fmt.Errorf("input %d invalid public key: %w", inputIndex, err)
		}

		if addr != referencedOutput.Address {
			return 0, fmt.Errorf("input %d public key does not match referenced output address", inputIndex)
		}

		if !tx.VerifyInputSignature(inputIndex) {
			return 0, fmt.Errorf("input %d signature verification failed", inputIndex)
		}

		inputSum += referencedOutput.Value
		spentInBlock[key] = struct{}{}
	}

	if inputSum < outputSum {
		return 0, errors.New("input sum is less than output sum")
	}

	return inputSum - outputSum, nil
}

func ValidateCoinbaseTransaction(tx Transaction, totalFees uint64) error {
	if !tx.IsCoinbase() {
		return errors.New("coinbase transaction must have no inputs")
	}

	if len(tx.Outputs) == 0 {
		return errors.New("coinbase transaction must have at least one output")
	}

	expectedID, err := tx.ComputeID()
	if err != nil {
		return fmt.Errorf("compute coinbase tx id: %w", err)
	}

	if tx.ID != expectedID {
		return errors.New("invalid coinbase transaction id")
	}

	var totalOutput uint64
	for _, out := range tx.Outputs {
		totalOutput += out.Value
	}

	expectedTotal := BlockReward + totalFees
	if totalOutput != expectedTotal {
		return fmt.Errorf("invalid coinbase total output: got %d want %d", totalOutput, expectedTotal)
	}

	return nil
}
