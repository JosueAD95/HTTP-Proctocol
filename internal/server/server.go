package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/JosueAD95/httpfromtcp/internal/request"
	"github.com/JosueAD95/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, r *request.Request)

type Server struct {
	listener net.Listener
	state    atomic.Bool //true if server is running
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := Server{
		listener: l,
		state:    atomic.Bool{},
		handler:  handler,
	}
	server.state.Store(true)

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	s.state.Store(false)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for s.state.Load() {
		// Wait for a connection.
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println(err)
		return
	}
	writer := response.Writer{
		Conn: conn,
	}

	s.handler(&writer, req)
}
