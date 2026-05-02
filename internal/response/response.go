package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Agustin-del/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

func WriteStatusLine(w io.Writer, sc StatusCode) error {
	var msg string
	switch sc {
	case OK:
		msg = "OK"
	case BAD_REQUEST:
		msg = "Bad Request"
	case INTERNAL_SERVER_ERROR:
		msg = "Internal Server Error"
	}

	if msg != "" {
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", sc, msg)
		return err
	}

	_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n", sc)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	hs := headers.NewHeaders()
	hs.Set("content-length", strconv.FormatUint(uint64(contentLen), 10))
	hs.Set("connection", "close")
	hs.Set("content-type", "text/plain")

	return *hs
}

func WriteHeaders(w io.Writer, hs headers.Headers) error {
	var e error
	hs.Iterate(func (k, v string){
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			e = err
		}
	})

	_, e = fmt.Fprintf(w, "\r\n")
	return e
}
