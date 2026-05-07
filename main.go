// HTTP From TCP
// main package
// main function
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	// Open File
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", inputFilePath, err)
	}

	fmt.Printf("Reading data from %s\n", inputFilePath)
	fmt.Println("==============================================")

	msgs := getLinesChannel(file)
	for msg := range msgs {
		fmt.Printf("read: %s\n", msg)
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
