package network

import (
	"fmt"
	"net"
)

func StartClient() {
	client, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	client.Write([]byte("I am CLIENT\n"))

	buf := make([]byte, 4096)

	n, err := client.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(buf[:n]))
}
