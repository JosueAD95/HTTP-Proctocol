package main

import (
	"fmt"
	"net"
	"os"

	"github.com/JosueAD95/httpfromtcp/internal/request"
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
		fmt.Println("Accepted connection from: ", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("Error parsing the request: ", err.Error())
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))

		conn.Close()
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
