package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/JosueAD95/httpfromtcp/internal/headers"
)

type Writer struct {
	Conn  io.Writer
	state HttpResponseState
}

type HttpResponseState int

const (
	statusLineState HttpResponseState = iota
	headerState
	bodyState
	endedState
)

type StatusCode int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

const CRLF = "\r\n"

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state: statusLineState,
		Conn:  w,
	}
}

func GetDefaultHeaders(contentLen int, contentType string) headers.Headers {
	defautlHeaders := headers.NewHeaders()
	defautlHeaders["Connection"] = "close"
	defautlHeaders["Content-Type"] = contentType
	defautlHeaders["Content-Length"] = strconv.Itoa(contentLen)

	return defautlHeaders
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != statusLineState {
		return fmt.Errorf("cannot write status line in state %d", w.state)
	}
	defer func() { w.state = headerState }()
	statusMap := map[StatusCode]string{
		StatusCodeSuccess:             "Ok",
		StatusCodeBadRequest:          "Bad Request",
		StatusCodeInternalServerError: "Internal Server Error",
	}
	value, _ := statusMap[statusCode]
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s%s", statusCode, value, CRLF)

	_, err := w.Conn.Write([]byte(statusLine))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != headerState {
		return fmt.Errorf("cannot write headers state %d", w.state)
	}
	defer func() { w.state = bodyState }()
	for key, value := range headers {
		_, err := w.Conn.Write([]byte(fmt.Sprintf("%s: %s%s", key, value, CRLF)))
		if err != nil {
			return err
		}
	}
	_, err := w.Conn.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(body []byte) error {
	if w.state != bodyState {
		return fmt.Errorf("cannot write body state %d", w.state)
	}
	defer func() { w.state = endedState }()

	_, err := w.Conn.Write(body)
	return err
}
