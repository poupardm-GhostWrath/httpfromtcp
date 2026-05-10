// TCP Listener
package main

import (
	"fmt"
	"log"
	"net"

	"github.com/poupardm-GhostWrath/httpfromtcp/internal/request"
)

const (
	port = ":42069"
)

func main() {
	// TCP listener
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %v\n", err)
	}
	defer listener.Close()

	fmt.Printf("Listening for TCP traffic on %s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HTTPVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s", key, value)
		}
		fmt.Printf("Connection to %s closed", conn.RemoteAddr().String())
	}
}
