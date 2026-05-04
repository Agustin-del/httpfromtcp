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

type writerState int

const (
	initial writerState = iota
	statusLineWritten
	headersWritten
	bodyWritten
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  initial,
	}
}

func (w *Writer) WriteStatusLine(sc StatusCode) error {
	if w.state != initial {
		return fmt.Errorf("status line already written or invalid order")
	}

	var msg string
	switch sc {
	case OK:
		msg = "OK"
	case BAD_REQUEST:
		msg = "Bad Request"
	case INTERNAL_SERVER_ERROR:
		msg = "Internal Server Error"
	}

	var err error
	if msg != "" {
		_, err = fmt.Fprintf(w.writer, "HTTP/1.1 %d %s\r\n", sc, msg)
	} else {
		_, err = fmt.Fprintf(w.writer, "HTTP/1.1 %d \r\n", sc)
	}

	if err == nil {
		w.state = statusLineWritten
	}
	return err
}

func (w *Writer) WriteHeaders(hs headers.Headers) error {
	if w.state != statusLineWritten {
		return fmt.Errorf("headers must be written after status line")
	}

	var err error
	hs.Iterate(func(k, v string) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
	})

	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w.writer, "\r\n")

	if err == nil {
		w.state = headersWritten
	}

	return err
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("body must be written after headers")
	}
	n, err := w.writer.Write(body)
	if err != nil {
		return 0, err
	}

	w.state = bodyWritten
	return n, err
}

func (w *Writer) WriteChunkedBody(data []byte) (int, error) {
	if w.state != headersWritten {
		return 0, fmt.Errorf("body must be written after headers")
	}
	totalBytesWrited := 0

	n1, err := fmt.Fprintf(w.writer, "%x\r\n", len(data))
	if err != nil {
		return totalBytesWrited, err
	}
	totalBytesWrited += n1
	n2, err := w.writer.Write(data)
	if err != nil {
		return totalBytesWrited, err
	}
	totalBytesWrited += n2

	n3, err := fmt.Fprintf(w.writer, "\r\n")

	if err != nil {
		return totalBytesWrited, err
	}
	totalBytesWrited += n3

	return totalBytesWrited, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.writer.Write([]byte("0\r\n"))

	if err == nil {
		w.state = bodyWritten
	}
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	var err error
	h.Iterate(func(k, v string) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
	})

	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w.writer, "\r\n")

	return err
}

func GetDefaultHeaders(contentLen int, chunk bool) headers.Headers {
	hs := headers.NewHeaders()
	if !chunk {
		hs.Set("content-length", strconv.FormatUint(uint64(contentLen), 10))
	} else {
		hs.Set("transfer-encoding", "chunked")
	}

	hs.Set("connection", "close")
	hs.Set("content-type", "text/plain")

	return *hs
}
