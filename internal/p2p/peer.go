package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type Peer struct {
	conn net.Conn
	mu   sync.Mutex
	addr string
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		conn: conn,
		addr: conn.RemoteAddr().String(),
	}
}

func (p *Peer) Addr() string {
	return p.addr
}

func (p *Peer) Send(msg Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	data = append(data, '\n')

	if _, err := p.conn.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func ReadMessage(scanner *bufio.Scanner, msg *Message) error {
	if !scanner.Scan() {
		return scanner.Err()
	}

	if err := json.Unmarshal(scanner.Bytes(), msg); err != nil {
		return fmt.Errorf("unmarshal message: %w", err)
	}

	return nil
}
