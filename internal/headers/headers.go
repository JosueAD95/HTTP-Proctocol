package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

type Headers map[string]string

const CRLF = "\r\n"

var VALIDCHARACTERS = []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	value, ok := h[key]
	return value, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 { // no CRLF found
		return 0, false, nil
	}

	if idx == 0 { // no more headers
		return 2, true, nil //adding 2 for the CRLF
	}

	header := data[:idx]
	headerLength := len(header)

	colonIdx := bytes.IndexByte(header, byte(':')) // -1 if the byte is not found
	if colonIdx <= 0 ||                            //validates if ':' exist or is at the beginning
		header[colonIdx-1] == byte(' ') || //has a space at the left of the :
		headerLength == colonIdx+1 { //it is at the end of the slice
		return 0, false, fmt.Errorf("malformed header: %s", string(header))
	}

	key := strings.ToLower(string(bytes.TrimSpace(header[:colonIdx])))
	if !isValidHeaderKey(key) {
		return 0, false, fmt.Errorf("invalid characters in header key: %s", key)
	}

	value := string(bytes.TrimSpace(header[colonIdx+1:]))
	if prevVal, exist := h[key]; exist { //When the header is used multiple times
		h[key] = prevVal + ", " + value
	} else {
		h[key] = value
	}

	return headerLength + 2, false, nil // add 2 counting for the CRLF
}

func isValidHeaderKey(key string) bool {
	for _, c := range key {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(r rune) bool {
	if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
		return true
	}
	return slices.Contains(VALIDCHARACTERS, r)
}
