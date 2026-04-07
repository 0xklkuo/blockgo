package node

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"blockgo/internal/blockchain"
	"blockgo/internal/config"
	"blockgo/internal/consensus"
	blockcrypto "blockgo/internal/crypto"
	"blockgo/internal/mempool"
	"blockgo/internal/p2p"
	"blockgo/internal/store"
)

const dbFileName = "blockgo.db"

type Node struct {
	mu sync.Mutex

	cfg          config.NodeConfig
	logger       *slog.Logger
	validatorSet *consensus.ValidatorSet
	mempool      *mempool.Pool
	store        *store.Store
	p2pServer    *p2p.Server

	head  *blockchain.Block
	utxos *blockchain.UTXOSet

	localPub []byte

	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func New(cfg config.NodeConfig, genesis *blockchain.Block, logger *slog.Logger) (*Node, error) {
	if err := cfg.NormalizeAndValidate(); err != nil {
		return nil, err
	}
	if genesis == nil {
		return nil, errors.New("nil genesis block")
	}
	if err := blockchain.ValidateGenesisBlock(genesis); err != nil {
		return nil, fmt.Errorf("invalid genesis block: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	set, err := consensus.NewValidatorSet(cfg.ValidatorPublicKeys)
	if err != nil {
		return nil, err
	}

	localPub := cfg.LocalPublicKey()
	if !set.IsValidator(localPub) {
		return nil, errors.New("local validator key is not in validator set")
	}

	dbPath := filepath.Join(cfg.DataDir, dbFileName)
	st, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	head, utxos, err := loadOrInitState(st, genesis)
	if err != nil {
		_ = st.Close()
		return nil, err
	}

	n := &Node{
		cfg:          cfg,
		logger:       logger,
		validatorSet: set,
		mempool:      mempool.New(),
		store:        st,
		head:         head,
		utxos:        utxos,
		localPub:     localPub,
		stopCh:       make(chan struct{}),
	}

	n.p2pServer = p2p.NewServer(logger, n)

	return n, nil
}

func loadOrInitState(st *store.Store, genesis *blockchain.Block) (*blockchain.Block, *blockchain.UTXOSet, error) {
	head, err := st.LoadHead()
	if err != nil {
		return nil, nil, fmt.Errorf("load head: %w", err)
	}

	if head != nil {
		utxos, err := st.LoadUTXOSet()
		if err != nil {
			return nil, nil, fmt.Errorf("load utxos: %w", err)
		}
		return head, utxos, nil
	}

	utxos := blockchain.NewUTXOSet()
	if err := utxos.ApplyGenesisBlock(genesis); err != nil {
		return nil, nil, fmt.Errorf("apply genesis: %w", err)
	}

	if err := st.SaveGenesis(genesis, utxos); err != nil {
		return nil, nil, fmt.Errorf("save genesis: %w", err)
	}

	return genesis, utxos, nil
}

func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return errors.New("node already running")
	}

	if err := n.p2pServer.Start(n.cfg.ListenAddr); err != nil {
		return fmt.Errorf("start p2p server: %w", err)
	}

	for _, peerAddr := range n.cfg.Peers {
		go func(addr string) {
			if err := n.p2pServer.Connect(addr); err != nil {
				n.logger.Warn("connect peer failed", "peer", addr, "error", err)
			}
		}(peerAddr)
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
		if n.store != nil {
			return n.store.Close()
		}
		return nil
	}

	close(n.stopCh)
	n.running = false
	n.mu.Unlock()

	n.wg.Wait()

	if err := n.p2pServer.Stop(); err != nil {
		return err
	}

	if n.store != nil {
		return n.store.Close()
	}

	return nil
}

func (n *Node) producerLoop() {
	defer n.wg.Done()

	ticker := time.NewTicker(n.cfg.BlockInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			block, err := n.TryProduceBlock(time.Now().UTC())
			if err != nil {
				n.logger.Warn("produce block failed", "error", err)
				continue
			}
			if block != nil {
				n.logger.Info("produced block", "height", block.Header.Height, "hash", block.Hash.String())
			}
		case <-n.stopCh:
			return
		}
	}
}

func (n *Node) SubmitTransaction(tx blockchain.Transaction) error {
	return n.submitTransaction(tx, true)
}

func (n *Node) submitTransaction(tx blockchain.Transaction, broadcast bool) error {
	if tx.IsCoinbase() {
		return errors.New("coinbase transaction cannot be submitted to mempool")
	}
	if blockchain.IsZeroHash(tx.ID) {
		return errors.New("transaction id is zero")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.mempool.Has(tx.ID) {
		return nil
	}

	working := n.utxos.Clone()
	_, err := blockchain.ValidateTransaction(tx, working, map[blockchain.UTXOKey]struct{}{})
	if err != nil {
		return fmt.Errorf("validate tx before mempool insert: %w", err)
	}

	err = n.mempool.Add(tx)
	if err != nil {
		if errors.Is(err, mempool.ErrTxAlreadyExists) {
			return nil
		}
		return err
	}

	if broadcast {
		n.p2pServer.Broadcast(p2p.Message{
			Type: p2p.MessageTypeNewTx,
			Tx:   &tx,
		})
	}

	return nil
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
		return nil, nil
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

	if err := n.store.SaveBlock(block, nextUTXOs); err != nil {
		return nil, fmt.Errorf("save block: %w", err)
	}

	n.head = block
	n.utxos = nextUTXOs
	for _, tx := range selected {
		n.mempool.Remove(tx.ID)
	}

	n.p2pServer.Broadcast(p2p.Message{
		Type:  p2p.MessageTypeNewBlock,
		Block: block,
	})

	return block, nil
}

func (n *Node) ApplyExternalBlock(block *blockchain.Block) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if block == nil {
		return errors.New("nil block")
	}

	if block.Header.Height <= n.head.Header.Height {
		return nil
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

	if err := n.store.SaveBlock(block, nextUTXOs); err != nil {
		return fmt.Errorf("save block: %w", err)
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

func (n *Node) LocalHello() p2p.Message {
	n.mu.Lock()
	defer n.mu.Unlock()

	return p2p.Message{
		Type:   p2p.MessageTypeHello,
		NodeID: n.cfg.NodeID,
		Height: n.head.Header.Height,
	}
}

func (n *Node) HandleMessage(peer *p2p.Peer, msg p2p.Message) error {
	switch msg.Type {
	case p2p.MessageTypeHello:
		if msg.Height > n.Head().Header.Height {
			return peer.Send(p2p.Message{
				Type:       p2p.MessageTypeGetBlocks,
				FromHeight: n.Head().Header.Height + 1,
			})
		}
		return nil

	case p2p.MessageTypeGetBlocks:
		return n.handleGetBlocks(peer, msg.FromHeight)

	case p2p.MessageTypeBlocks:
		for i := range msg.Blocks {
			block := msg.Blocks[i]
			if err := n.ApplyExternalBlock(&block); err != nil {
				return fmt.Errorf("apply synced block at height %d: %w", block.Header.Height, err)
			}
		}
		return nil

	case p2p.MessageTypeNewTx:
		if msg.Tx == nil {
			return errors.New("new_tx missing transaction")
		}
		return n.submitTransaction(*msg.Tx, false)

	case p2p.MessageTypeNewBlock:
		if msg.Block == nil {
			return errors.New("new_block missing block")
		}
		return n.ApplyExternalBlock(msg.Block)

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (n *Node) handleGetBlocks(peer *p2p.Peer, fromHeight uint64) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if fromHeight > n.head.Header.Height {
		return nil
	}

	blocks := make([]blockchain.Block, 0, n.head.Header.Height-fromHeight+1)
	current := n.head

	for current != nil && current.Header.Height >= fromHeight {
		blocks = append([]blockchain.Block{*current}, blocks...)
		if current.Header.Height == 0 {
			break
		}

		prev, err := n.store.LoadBlock(current.Header.PrevBlockHash)
		if err != nil {
			return fmt.Errorf("load previous block: %w", err)
		}
		current = prev
	}

	return peer.Send(p2p.Message{
		Type:   p2p.MessageTypeBlocks,
		Blocks: blocks,
	})
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
