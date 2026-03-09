package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/JosueAD95/httpfromtcp/internal/response"
)

func WriteFromHttpBin(w *response.Writer, resource string) {
	fullpath := fmt.Sprintf("https://httpbin.org/stream/%s", resource)
	res, err := http.Get(fullpath)
	if err != nil {
		log.Fatal(err)
		//TODO: Add error handling
		return
	}
	buffer := make([]byte, 32)
	defer res.Body.Close()
	if err = w.WriteStatusLine(response.StatusCodeSuccess); err != nil {
		log.Fatal(err)
		//TODO: Add error handling
		return

	}
	headers := response.GetDefaultHeaders(0, res.Header.Get("Content-Type"))
	delete(headers, "Content-Length")
	headers["Transfer-Encoding"] = "chunked"
	if err = w.WriteHeaders(headers); err != nil {
		log.Fatal(err)
		//TODO: Add error handling
		return
	}

	for {
		bytesRead, err := res.Body.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				w.WriteChunkedBodyDone()
				break
			}
			log.Fatal(err)
			//TODO: Add error handling
			return
		}
		log.Printf("%d bytes read from HTTPBin\n", bytesRead)
		bytesWrote, err := w.WriteChunkedBody(buffer)
		if err != nil {
			log.Fatal(err)
			//TODO: Add error handling
			return
		}
		log.Printf("%d bytes read from HTTPBin\n", bytesWrote)
	}
}
