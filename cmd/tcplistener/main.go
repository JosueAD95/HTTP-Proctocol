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

		reqLine := fmt.Sprintf(`
Request line:
- Method: %s
- Target: %s
- Version: %s`,
			req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Println(reqLine)

		conn.Close()
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
