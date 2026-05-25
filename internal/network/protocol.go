package network

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/Suryakiran46/lantern/internal/config"
)

func WriteMessage(conn net.Conn, msg config.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func ReadMessage(reader *bufio.Reader) (config.Message, error) {
	data, err := reader.ReadString('\n')
	if err != nil {
		return config.Message{}, err
	}
	var message config.Message
	err = json.Unmarshal([]byte(data), &message)
	if err != nil {
		return config.Message{}, err
	}
	return message, nil
}
