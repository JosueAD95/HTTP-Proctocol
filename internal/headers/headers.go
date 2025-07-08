package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

const CRLF = "\r\n"

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 { // no CRLF found
		return 0, false, nil
	}

	header := data[:idx]
	headerLength := len(header)
	if headerLength == 0 { // no more headers
		return 0, true, nil
	}

	colonIdx := bytes.IndexByte(header, byte(':')) // -1 if the byte is not found
	if colonIdx <= 0 ||                            //validates if ':' exist or is at the beginning
		header[colonIdx-1] == byte(' ') || //has a space at the left
		headerLength == colonIdx+1 { //it is at the end of the slice
		return 0, false, fmt.Errorf("malformed header: %s", string(header))
	}

	key := string(bytes.TrimSpace(header[:colonIdx]))
	value := string(bytes.TrimSpace(header[colonIdx+1:]))
	h[key] = value

	return headerLength + 2, false, nil // add 2 counting for the CRLF
}
