// TCP Listener
package main

import (
	"errors"
	"fmt"
	"io"
	"log"

	// "os"
	"net"
	"strings"
)

// const inputFilePath = "messages.txt"
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
		msgs := getLinesChannel(conn)
		for msg := range msgs {
			fmt.Println(msg)
		}
		fmt.Printf("Connection to %s closed", conn.RemoteAddr().String())
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)
		line := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				if line != "" {
					ch <- line
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %v\n", err)
				return
			}
			str := string(data[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				ch <- fmt.Sprintf("%s%s", line, parts[i])
				line = ""
			}
			line += parts[len(parts)-1]
		}
	}()
	return ch
}
