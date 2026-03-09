package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/JosueAD95/httpfromtcp/internal/proxy"
	"github.com/JosueAD95/httpfromtcp/internal/request"
	"github.com/JosueAD95/httpfromtcp/internal/response"
	"github.com/JosueAD95/httpfromtcp/internal/server"
)

const port = 42069

const successHtml = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

const errorHtml = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalErrorHtml = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const httpBin = "/httpbin/"

func handler(w *response.Writer, r *request.Request) {
	var body []byte
	path := r.RequestLine.RequestTarget
	switch {
	case path == "/yourproblem":
		if err := w.WriteStatusLine(response.StatusCodeBadRequest); err != nil {
			return
		}
		body = []byte(errorHtml)
	case path == "/myproblem":
		if err := w.WriteStatusLine(response.StatusCodeInternalServerError); err != nil {
			return
		}
		body = []byte(internalErrorHtml)
	case strings.HasPrefix(path, httpBin):
		resource := strings.TrimPrefix(path, httpBin)
		proxy.WriteFromHttpBin(w, resource)
		return
	default:
		if err := w.WriteStatusLine(response.StatusCodeSuccess); err != nil {
			return
		}
		body = []byte(successHtml)
	}
	headers := response.GetDefaultHeaders(len(body), "text/html")
	if err := w.WriteHeaders(headers); err != nil {
		return
	}
	w.WriteBody(body)
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
