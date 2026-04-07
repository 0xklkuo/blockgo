package mempool

import (
	"errors"
	"sort"
	"sync"

	"blockgo/internal/blockchain"
)

var ErrTxAlreadyExists = errors.New("transaction already exists in mempool")

type Pool struct {
	mu  sync.RWMutex
	txs map[string]blockchain.Transaction
}

func New() *Pool {
	return &Pool{
		txs: make(map[string]blockchain.Transaction),
	}
}

func (p *Pool) Add(tx blockchain.Transaction) error {
	if tx.IsCoinbase() {
		return errors.New("coinbase transaction is not allowed in mempool")
	}
	if blockchain.IsZeroHash(tx.ID) {
		return errors.New("transaction id is zero")
	}

	key := tx.ID.String()

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.txs[key]; exists {
		return ErrTxAlreadyExists
	}

	p.txs[key] = tx
	return nil
}

func (p *Pool) Has(txID blockchain.Hash) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, ok := p.txs[txID.String()]
	return ok
}

func (p *Pool) Remove(txID blockchain.Hash) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.txs, txID.String())
}

func (p *Pool) Snapshot() []blockchain.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]string, 0, len(p.txs))
	for k := range p.txs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]blockchain.Transaction, 0, len(keys))
	for _, k := range keys {
		out = append(out, p.txs[k])
	}
	return out
}

func (p *Pool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.txs)
}
