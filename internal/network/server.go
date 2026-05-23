package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/Suryakiran46/lantern/internal/config"
)

type Server struct {
	listener    net.Listener
	connections map[string]net.Conn
	mu          sync.RWMutex
	msgChan     chan config.Message
	peerChan    chan config.PeerEvent
	port        int
}

func NewServer(p int) *Server {
	connections := make(map[string]net.Conn)
	msg := make(chan config.Message, 100)
	peer := make(chan config.PeerEvent, 10)
	s := Server{
		connections: connections,
		msgChan:     msg,
		peerChan:    peer,
		port:        p,
	}
	return &s
}

func (s *Server) Listen() error {
	port := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(s)
	return nil
}

func acceptLoop(s *Server) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go handleConn(s, conn)
	}
}

func handleConn(s *Server, conn net.Conn) {
	address := conn.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(address)
	s.mu.Lock()
	s.connections[ip] = conn
	s.mu.Unlock()
	s.peerChan <- config.PeerEvent{
		Type: config.PeerConnected,
		Peer: ip,
	}
	go readLoop(s, conn, ip)
}

func (s *Server) Send(to string, msg config.Message) error {
	s.mu.RLock()
	conn := s.connections[to]
	s.mu.RUnlock()
	if conn == nil {
		return fmt.Errorf("Peer %s not Connected", to)
	}
	return WriteMessage(conn, msg)
}

func (s *Server) Broadcast(msg config.Message) error {
	s.mu.RLock()
	for _, conn := range s.connections {
		WriteMessage(conn, msg)
	}
	s.mu.RUnlock()
	return nil
}

func (s *Server) MessageChan() <-chan config.Message {
	return s.msgChan
}

func (s *Server) PeerChan() <-chan config.PeerEvent {
	return s.peerChan
}

func (s *Server) Close() {
	s.listener.Close()
	s.mu.Lock()
	for _, conn := range s.connections {
		conn.Close()
	}
	s.mu.Unlock()
}
