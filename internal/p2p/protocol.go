package p2p

import "blockgo/internal/blockchain"

const (
	MessageTypeHello     = "hello"
	MessageTypeGetBlocks = "get_blocks"
	MessageTypeBlocks    = "blocks"
	MessageTypeNewTx     = "new_tx"
	MessageTypeNewBlock  = "new_block"
)

type Message struct {
	Type       string                  `json:"type"`
	NodeID     string                  `json:"node_id,omitempty"`
	Height     uint64                  `json:"height,omitempty"`
	FromHeight uint64                  `json:"from_height,omitempty"`
	Block      *blockchain.Block       `json:"block,omitempty"`
	Blocks     []blockchain.Block      `json:"blocks,omitempty"`
	Tx         *blockchain.Transaction `json:"tx,omitempty"`
}
