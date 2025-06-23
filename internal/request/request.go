package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

var SUPPORTEDMETHODS = [4]string{"GET", "PUT", "POST", "DELETE"}

const BUFFERSIZE = 8

const CRLF = "\r\n"

type requestState int

const (
	requestInitialized requestState = iota
	requestDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 { //just need more data
		return 0, nil
	}
	switch r.state {
	case requestInitialized:
		requestLine, _, err := parseRequestLine(data[:idx])
		if err != nil { //something actually went wrong
			return 0, err
		}
		r.RequestLine = *requestLine
		r.state = requestDone
		return idx + 2, nil
	case requestDone:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{
		state: requestInitialized,
	}
	var readToIndex int
	var bytesRead int
	var bytesParsed int
	var err error
	buffer := make([]byte, BUFFERSIZE, BUFFERSIZE)

	for req.state != requestDone {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		bytesRead, err = reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = requestDone
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesParsed, err = req.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buffer, buffer[bytesParsed:])
		readToIndex += bytesParsed
	}
	return &req, err
}

func parseRequestLine(reqLine []byte) (*RequestLine, int, error) {
	lineParts := strings.Split(string(reqLine), " ")
	if len(lineParts) != 3 {
		return nil, 0, fmt.Errorf("incorrect format in request-line: %s", string(reqLine))
	}

	if !isMethodSupported(lineParts[0]) {
		return nil, 0, fmt.Errorf("method is not supported: %s", lineParts[0])
	}

	if lineParts[2] != "HTTP/1.1" {
		return nil, 0, fmt.Errorf("unsupported HTTP version: %s", lineParts[2])
	}
	return &RequestLine{
		Method:        lineParts[0],
		RequestTarget: lineParts[1],
		HttpVersion:   lineParts[2],
	}, 0, nil
}

func isMethodSupported(method string) bool {
	for _, supportedMethod := range SUPPORTEDMETHODS {
		if method == supportedMethod {
			return true
		}
	}
	return false
}
