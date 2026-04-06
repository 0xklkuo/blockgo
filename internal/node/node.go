package node

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"blockgo/internal/blockchain"
	"blockgo/internal/config"
	"blockgo/internal/consensus"
	blockcrypto "blockgo/internal/crypto"
	"blockgo/internal/mempool"
)

type Node struct {
	mu sync.Mutex

	cfg          config.NodeConfig
	validatorSet *consensus.ValidatorSet
	mempool      *mempool.Pool

	head  *blockchain.Block
	utxos *blockchain.UTXOSet

	localPub []byte

	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func New(cfg config.NodeConfig, genesis *blockchain.Block) (*Node, error) {
	if err := cfg.NormalizeAndValidate(); err != nil {
		return nil, err
	}
	if genesis == nil {
		return nil, errors.New("nil genesis block")
	}
	if err := blockchain.ValidateGenesisBlock(genesis); err != nil {
		return nil, fmt.Errorf("invalid genesis block: %w", err)
	}

	set, err := consensus.NewValidatorSet(cfg.ValidatorPublicKeys)
	if err != nil {
		return nil, err
	}

	localPub := cfg.LocalPublicKey()
	if !set.IsValidator(localPub) {
		return nil, errors.New("local validator key is not in validator set")
	}

	utxos := blockchain.NewUTXOSet()
	if err := utxos.ApplyGenesisBlock(genesis); err != nil {
		return nil, fmt.Errorf("apply genesis to utxos: %w", err)
	}

	return &Node{
		cfg:          cfg,
		validatorSet: set,
		mempool:      mempool.New(),
		head:         genesis,
		utxos:        utxos,
		localPub:     localPub,
		stopCh:       make(chan struct{}),
	}, nil
}

func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return errors.New("node already running")
	}

	n.running = true
	n.stopCh = make(chan struct{})

	n.wg.Add(1)
	go n.producerLoop()

	return nil
}

func (n *Node) Stop() error {
	n.mu.Lock()
	if !n.running {
		n.mu.Unlock()
		return nil
	}
	close(n.stopCh)
	n.running = false
	n.mu.Unlock()

	n.wg.Wait()
	return nil
}

func (n *Node) producerLoop() {
	defer n.wg.Done()

	ticker := time.NewTicker(n.cfg.BlockInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, _ = n.TryProduceBlock(time.Now().UTC())
		case <-n.stopCh:
			return
		}
	}
}

func (n *Node) SubmitTransaction(tx blockchain.Transaction) error {
	if tx.IsCoinbase() {
		return errors.New("coinbase transaction cannot be submitted to mempool")
	}
	if blockchain.IsZeroHash(tx.ID) {
		return errors.New("transaction id is zero")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	working := n.utxos.Clone()
	_, err := blockchain.ValidateTransaction(tx, working, map[blockchain.UTXOKey]struct{}{})
	if err != nil {
		return fmt.Errorf("validate tx before mempool insert: %w", err)
	}

	return n.mempool.Add(tx)
}

func (n *Node) TryProduceBlock(now time.Time) (*blockchain.Block, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	nextHeight := n.head.Header.Height + 1

	expectedProposer, err := n.validatorSet.Proposer(nextHeight)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(expectedProposer, n.localPub) {
		return nil, nil // not our turn
	}

	selected, totalFees := n.selectTransactionsLocked()

	rewardAddress, err := blockcrypto.AddressFromPublicKey(n.localPub)
	if err != nil {
		return nil, fmt.Errorf("derive reward address: %w", err)
	}

	coinbase := blockchain.Transaction{
		Outputs: []blockchain.TxOutput{
			{
				Value:   blockchain.BlockReward + totalFees,
				Address: rewardAddress,
			},
		},
	}
	if err := coinbase.Finalize(); err != nil {
		return nil, fmt.Errorf("finalize coinbase: %w", err)
	}

	txs := make([]blockchain.Transaction, 0, 1+len(selected))
	txs = append(txs, coinbase)
	txs = append(txs, selected...)

	block, err := blockchain.NewBlock(nextHeight, n.head.Hash, n.localPub, txs, now.UTC())
	if err != nil {
		return nil, fmt.Errorf("build block: %w", err)
	}

	if err := n.validatorSet.SignBlock(block, n.cfg.LocalPrivateKey); err != nil {
		return nil, fmt.Errorf("sign block: %w", err)
	}

	if err := n.validatorSet.ValidateBlock(block); err != nil {
		return nil, fmt.Errorf("consensus validation failed: %w", err)
	}

	if err := blockchain.ValidateBlock(block, n.head, n.utxos); err != nil {
		return nil, fmt.Errorf("blockchain validation failed: %w", err)
	}

	nextUTXOs := n.utxos.Clone()
	if err := nextUTXOs.ApplyBlock(block); err != nil {
		return nil, fmt.Errorf("apply block: %w", err)
	}

	n.head = block
	n.utxos = nextUTXOs
	for _, tx := range selected {
		n.mempool.Remove(tx.ID)
	}

	return block, nil
}

func (n *Node) ApplyExternalBlock(block *blockchain.Block) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if block == nil {
		return errors.New("nil block")
	}

	if err := n.validatorSet.ValidateBlock(block); err != nil {
		return fmt.Errorf("consensus validation failed: %w", err)
	}

	if err := blockchain.ValidateBlock(block, n.head, n.utxos); err != nil {
		return fmt.Errorf("blockchain validation failed: %w", err)
	}

	nextUTXOs := n.utxos.Clone()
	if err := nextUTXOs.ApplyBlock(block); err != nil {
		return fmt.Errorf("apply block: %w", err)
	}

	n.head = block
	n.utxos = nextUTXOs

	for _, tx := range block.Transactions {
		n.mempool.Remove(tx.ID)
	}

	return nil
}

func (n *Node) Head() *blockchain.Block {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.head
}

func (n *Node) MempoolLen() int {
	return n.mempool.Len()
}

func (n *Node) selectTransactionsLocked() ([]blockchain.Transaction, uint64) {
	pending := n.mempool.Snapshot()
	if len(pending) == 0 {
		return nil, 0
	}

	working := n.utxos.Clone()
	spent := map[blockchain.UTXOKey]struct{}{}

	selected := make([]blockchain.Transaction, 0, len(pending))
	var totalFees uint64

	for _, tx := range pending {
		if len(selected) >= n.cfg.MaxTxPerBlock {
			break
		}

		fee, err := blockchain.ValidateTransaction(tx, working, spent)
		if err != nil {
			continue
		}

		if err := blockchain.ApplyTransaction(working, tx); err != nil {
			continue
		}

		selected = append(selected, tx)
		totalFees += fee
	}

	return selected, totalFees
}
