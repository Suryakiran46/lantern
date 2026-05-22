package network

import (
	"bufio"
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
}

func NewServer() *Server {
	connections := make(map[string]net.Conn)
	msg := make(chan config.Message, 100)
	peer := make(chan config.PeerEvent, 10)
	s := Server{
		connections: connections,
		msgChan:     msg,
		peerChan:    peer,
	}
	return &s
}

func (s *Server) Listen(port int) error {
	p := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", p)
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
	ip := conn.RemoteAddr().String()
	s.mu.Lock()
	s.connections[ip] = conn
	s.mu.Unlock()
	s.peerChan <- config.PeerEvent{
		Type: config.PeerConnected,
		Peer: ip,
	}

	defer func() {
		s.mu.Lock()
		delete(s.connections, ip)
		s.mu.Unlock()
		s.peerChan <- config.PeerEvent{
			Type: config.PeerDisconnected,
			Peer: ip,
		}
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	for {
		msg, err := ReadMessage(reader)
		if err != nil {
			return
		}

		s.msgChan <- msg
	}
}

func (s *Server) MessageChan() <-chan config.Message {
	return s.msgChan
}

func (s *Server) PeerChan() <-chan config.PeerEvent {
	return s.peerChan
}

func (s *Server) Close() {
	s.listener.Close()
}
