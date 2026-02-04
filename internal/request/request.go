package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	headers "github.com/JosueAD95/httpfromtcp/internal/headers"
)

var SUPPORTEDMETHODS = [4]string{"GET", "PUT", "POST", "DELETE"}

var contentLength int

const BUFFERSIZE = 8

const CRLF = "\r\n"

type requestState int

const (
	requestInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 && r.state != requestStateParsingBody { //just need more data and is no Parsing the body
		return 0, nil
	}
	switch r.state {
	case requestInitialized:
		requestLine, _, err := parseRequestLine(data[:idx])
		if err != nil { //something actually went wrong
			return 0, err
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return idx + 2, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil { //something went wrong reading the headers
			return 0, err
		}
		if done {
			if contentLengthStr, ok := r.Headers.Get("Content-Length"); ok {
				contentLength, err = strconv.Atoi(contentLengthStr)
				if err != nil {
					return 0, fmt.Errorf("header Content-Length is not a valid number: %s", contentLengthStr)
				}
				r.state = requestStateParsingBody
			} else {
				r.state = requestStateDone
			}
		}
		return n, nil
	case requestStateParsingBody:
		r.Body = append(r.Body, data...)
		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("body length is greader than content-length header, %d > %d", len(r.Body), contentLength)
		}
		if len(r.Body) == contentLength {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, errors.New("reading data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{
		state:   requestInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
	var readToIndex, bytesRead, bytesParsed int
	var err error

	buffer := make([]byte, BUFFERSIZE, BUFFERSIZE)

	for req.state != requestStateDone {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		bytesRead, err = reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.state, bytesRead)
				}
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
		readToIndex -= bytesParsed
	}

	return &req, nil
}

func parseRequestLine(reqLine []byte) (*RequestLine, int, error) {
	lineParts := strings.Split(string(reqLine), " ")
	if len(lineParts) != 3 {
		return nil, 0, fmt.Errorf("incorrect format in request-line: %s", string(reqLine))
	}

	if !isMethodSupported(lineParts[0]) {
		return nil, 0, fmt.Errorf("method is not supported: %s", lineParts[0])
	}

	version := strings.TrimPrefix(lineParts[2], "HTTP/")
	if version != "1.1" {
		return nil, 0, fmt.Errorf("unsupported HTTP version: %s", lineParts[2])
	}

	return &RequestLine{
		Method:        lineParts[0],
		RequestTarget: lineParts[1],
		HttpVersion:   version,
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
