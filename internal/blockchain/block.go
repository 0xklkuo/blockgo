package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type BlockHeader struct {
	Height        uint64 `json:"height"`
	PrevBlockHash Hash   `json:"prev_block_hash"`
	MerkleRoot    Hash   `json:"merkle_root"`
	TimestampUTC  int64  `json:"timestamp_utc"`
	Validator     []byte `json:"validator"`
}

type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
	Signature    []byte        `json:"signature"`
	Hash         Hash          `json:"hash"`
}

type blockHashPayload struct {
	Header BlockHeader `json:"header"`
}

func (b Block) HashPayload() ([]byte, error) {
	payload := blockHashPayload{
		Header: b.Header,
	}

	return json.Marshal(payload)
}

func (b Block) ComputeHash() (Hash, error) {
	var zero Hash

	payload, err := b.HashPayload()
	if err != nil {
		return zero, fmt.Errorf("marshal block hash payload: %w", err)
	}

	sum := sha256.Sum256(payload)
	return Hash(sum), nil
}

func (b *Block) Finalize() error {
	if b == nil {
		return fmt.Errorf("nil block")
	}

	for i := range b.Transactions {
		if IsZeroHash(b.Transactions[i].ID) {
			if err := b.Transactions[i].Finalize(); err != nil {
				return fmt.Errorf("finalize tx %d: %w", i, err)
			}
		}
	}

	b.Header.MerkleRoot = MerkleRoot(b.Transactions)

	hash, err := b.ComputeHash()
	if err != nil {
		return err
	}

	b.Hash = hash
	return nil
}

func NewBlock(height uint64, prevHash Hash, validator []byte, txs []Transaction, timestamp time.Time) (*Block, error) {
	block := &Block{
		Header: BlockHeader{
			Height:        height,
			PrevBlockHash: prevHash,
			TimestampUTC:  timestamp.UTC().Unix(),
			Validator:     append([]byte(nil), validator...),
		},
		Transactions: append([]Transaction(nil), txs...),
	}

	if err := block.Finalize(); err != nil {
		return nil, err
	}

	return block, nil
}
