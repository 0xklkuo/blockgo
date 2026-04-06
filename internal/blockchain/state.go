package blockchain

import (
	"errors"
	"fmt"
	"maps"
)

type UTXOKey struct {
	TxID  Hash
	Index uint32
}

type UTXO struct {
	OutPoint OutPoint
	Output   TxOutput
}

type UTXOSet struct {
	entries map[UTXOKey]TxOutput
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		entries: make(map[UTXOKey]TxOutput),
	}
}

func (s *UTXOSet) Clone() *UTXOSet {
	if s == nil {
		return nil
	}

	out := NewUTXOSet()
	maps.Copy(out.entries, s.entries)
	return out
}

func (s *UTXOSet) Entries() map[UTXOKey]TxOutput {
	if s == nil {
		return nil
	}

	out := make(map[UTXOKey]TxOutput, len(s.entries))
	maps.Copy(out, s.entries)
	return out
}

func (s *UTXOSet) Get(out OutPoint) (TxOutput, bool) {
	if s == nil {
		return TxOutput{}, false
	}

	v, ok := s.entries[UTXOKey{
		TxID:  out.TxID,
		Index: out.Index,
	}]
	return v, ok
}

func (s *UTXOSet) Put(out OutPoint, output TxOutput) {
	if s == nil {
		return
	}

	s.entries[UTXOKey{
		TxID:  out.TxID,
		Index: out.Index,
	}] = output
}

func (s *UTXOSet) Delete(out OutPoint) {
	if s == nil {
		return
	}

	delete(s.entries, UTXOKey{
		TxID:  out.TxID,
		Index: out.Index,
	})
}

func (s *UTXOSet) Size() int {
	if s == nil {
		return 0
	}

	return len(s.entries)
}

func (s *UTXOSet) ApplyGenesisBlock(block *Block) error {
	if block == nil {
		return errors.New("nil block")
	}

	if block.Header.Height != 0 {
		return errors.New("genesis block height must be 0")
	}

	if len(block.Transactions) != 1 {
		return errors.New("genesis block must contain exactly one transaction")
	}

	tx := block.Transactions[0]
	if !tx.IsCoinbase() {
		return errors.New("genesis transaction must be coinbase-like")
	}

	if IsZeroHash(tx.ID) {
		return errors.New("genesis transaction id is zero")
	}

	for i, out := range tx.Outputs {
		s.Put(OutPoint{
			TxID:  tx.ID,
			Index: uint32(i),
		}, out)
	}

	return nil
}

func (s *UTXOSet) ApplyBlock(block *Block) error {
	if s == nil {
		return errors.New("nil utxo set")
	}

	if block == nil {
		return errors.New("nil block")
	}

	if len(block.Transactions) == 0 {
		return errors.New("block must contain at least one transaction")
	}

	if block.Header.Height == 0 {
		return s.ApplyGenesisBlock(block)
	}

	working := s.Clone()

	for txIndex, tx := range block.Transactions {
		if txIndex == 0 {
			if !tx.IsCoinbase() {
				return errors.New("first transaction must be coinbase")
			}
		} else if tx.IsCoinbase() {
			return errors.New("only first transaction may be coinbase")
		}

		if err := ApplyTransaction(working, tx); err != nil {
			return fmt.Errorf("apply tx %d: %w", txIndex, err)
		}
	}

	s.entries = working.entries
	return nil
}

func ApplyTransaction(utxos *UTXOSet, tx Transaction) error {
	if utxos == nil {
		return errors.New("nil utxo set")
	}

	if IsZeroHash(tx.ID) {
		return errors.New("transaction id is zero")
	}

	if tx.IsCoinbase() {
		for i, out := range tx.Outputs {
			utxos.Put(OutPoint{
				TxID:  tx.ID,
				Index: uint32(i),
			}, out)
		}
		return nil
	}

	for _, in := range tx.Inputs {
		if _, ok := utxos.Get(in.PrevOut); !ok {
			return fmt.Errorf("referenced utxo not found: %s:%d", in.PrevOut.TxID.String(), in.PrevOut.Index)
		}
	}

	for _, in := range tx.Inputs {
		utxos.Delete(in.PrevOut)
	}

	for i, out := range tx.Outputs {
		utxos.Put(OutPoint{
			TxID:  tx.ID,
			Index: uint32(i),
		}, out)
	}

	return nil
}
