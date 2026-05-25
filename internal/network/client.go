package network

import (
	"bufio"
	"fmt"
	"net"

	"github.com/Suryakiran46/lantern/internal/config"
)

func (s *Server) Connect(ip string) error {
	address := net.JoinHostPort(ip, fmt.Sprintf("%d", s.port))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.connections[ip] = conn
	s.mu.Unlock()

	s.peerChan <- config.PeerEvent{
		Type: config.PeerConnected,
		Peer: ip,
	}

	pongChan := make(chan struct{}, 1)
	go readLoop(s, conn, ip, pongChan)
	return nil
}

func readLoop(s *Server, conn net.Conn, ip string, pongChan chan struct{}) {
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
		//KeepAlive
		if msg.Type == "ping" {
			WriteMessage(conn, config.Message{
				Type: "pong",
			})
			continue
		} else if msg.Type == "pong" {
			select {
			case pongChan <- struct{}{}:
			default:
			}
			continue
		}

		s.msgChan <- msg
	}
}
