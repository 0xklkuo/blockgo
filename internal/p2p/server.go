package p2p

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
)

type Handler interface {
	HandleMessage(peer *Peer, msg Message) error
	LocalHello() Message
}

type Server struct {
	logger   *slog.Logger
	handler  Handler
	listener net.Listener

	mu    sync.Mutex
	peers map[string]*Peer
}

func NewServer(logger *slog.Logger, handler Handler) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	return &Server{
		logger:  logger,
		handler: handler,
		peers:   make(map[string]*Peer),
	}
}

func (s *Server) Start(listenAddr string) error {
	if listenAddr == "" {
		return errors.New("listen address is required")
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.listener = ln

	go s.acceptLoop()
	return nil
}

func (s *Server) Stop() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

func (s *Server) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial peer: %w", err)
	}

	peer := NewPeer(conn)
	s.addPeer(peer)

	if err := peer.Send(s.handler.LocalHello()); err != nil {
		_ = conn.Close()
		return err
	}

	go s.readLoop(peer)
	return nil
}

func (s *Server) Broadcast(msg Message) {
	s.mu.Lock()
	peers := make([]*Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.mu.Unlock()

	for _, peer := range peers {
		if err := peer.Send(msg); err != nil {
			s.logger.Warn("broadcast failed", "peer", peer.Addr(), "error", err)
		}
	}
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		peer := NewPeer(conn)
		s.addPeer(peer)

		if err := peer.Send(s.handler.LocalHello()); err != nil {
			s.logger.Warn("send hello failed", "peer", peer.Addr(), "error", err)
			_ = conn.Close()
			continue
		}

		go s.readLoop(peer)
	}
}

func (s *Server) readLoop(peer *Peer) {
	scanner := bufio.NewScanner(peer.conn)

	for {
		var msg Message
		if err := ReadMessage(scanner, &msg); err != nil {
			s.logger.Debug("peer read loop ended", "peer", peer.Addr(), "error", err)
			s.removePeer(peer)
			_ = peer.conn.Close()
			return
		}

		if err := s.handler.HandleMessage(peer, msg); err != nil {
			s.logger.Warn("handle message failed", "peer", peer.Addr(), "type", msg.Type, "error", err)
		}
	}
}

func (s *Server) addPeer(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[peer.Addr()] = peer
}

func (s *Server) removePeer(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peers, peer.Addr())
}
