package network

import (
	"fmt"
	"net"
)

func StartServer() {
	serv, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
	defer serv.Close()

	fmt.Println("Server listening on 8080")

	for {
		conn, err := serv.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			break
		}

		fmt.Printf("From client: %s", string(buf[:n]))

		conn.Write([]byte("I am SERVER\n"))
	}

	fmt.Println("client disconnected")
}
