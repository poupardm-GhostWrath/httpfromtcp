// UDP Sender
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	serverAddr = "localhost:42069"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatalf("error resolving UDP address: %v\n", err)
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("error dialing UDP network: %v\n", err)
	}
	defer udpConn.Close()

	fmt.Printf("Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n", serverAddr)

	fileReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		str, err := fileReader.ReadString('\n')
		if err != nil {
			log.Fatalf("error: %v\n", err)
			return
		}
		_, err = udpConn.Write([]byte(str))
		if err != nil {
			log.Fatalf("error writing to UDP: %v\n", err)
		}
		fmt.Printf("message sent: %s\n", str)
	}
}
