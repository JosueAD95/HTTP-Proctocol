package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:42069")
	if err != nil {
		fmt.Printf("Error starting to listen: %s", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Accepted connection from ", conn.RemoteAddr())

		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}

func getLinesChannel(file io.ReadCloser) <-chan string {
	lineChan := make(chan string)

	go func() {
		defer file.Close()

		buffer := make([]byte, 8, 8)
		var line string
		parts := make([]string, 2)
		for {
			n, err := file.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					lineChan <- line
					break
				}
				panic(fmt.Sprintf("Error reading file: %s", err.Error()))
			}

			parts = strings.Split(string(buffer[:n]), "\n")
			line += parts[0]
			if len(parts) >= 2 && parts[1] != "" {
				lineChan <- line
				line = parts[1]
			}
		}
		file.Close()
		close(lineChan)
	}()

	return lineChan
}
