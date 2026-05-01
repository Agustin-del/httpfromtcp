package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"github.com/Agustin-del/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	state       state
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type state int

const (
	initialized state = iota
	parsingHeaders
	parsingBody
	done
)

var lineSeparator = []byte("\r\n")

const bufferSize = 8

func newRequest() *Request {
	return &Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
		Body: make([]byte, 0),
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buffer := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != done {
		lenBufferActual := len(buffer)
		if lenBufferActual == readToIndex {
			newBuf := make([]byte, lenBufferActual*2)
			copy(newBuf, buffer[:readToIndex])
			buffer = newBuf
		}

		numBytesReaded, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if err == io.EOF {
				if request.state == parsingBody {
					return nil, errors.New("error: body not complete")	
				}
				request.state = done
				break
			}
			return nil, err
		}

		readToIndex += numBytesReaded
		numBytesParsed, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}

		if numBytesParsed > 0 {

			copy(buffer, buffer[numBytesParsed:readToIndex])
			readToIndex -= numBytesParsed
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case initialized:
		reqLine, consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}

		r.RequestLine = *reqLine
		r.state = parsingHeaders

		return consumed, nil

	case parsingHeaders:
		consumed, dne, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if consumed == 0 {
			return 0, nil
		}

		if dne {
			if _, ok := r.Headers.Get("content-length"); ok {
				r.state = parsingBody
			} else {
				r.state = done
			}

			return consumed, nil
		}

		return consumed, nil

	case parsingBody:
		cLengthString, _ := r.Headers.Get("content-length")

		cLength, err := strconv.Atoi(cLengthString)
		if err != nil {
			return 0, err
		}

		remaining := cLength - len(r.Body)	

		if len(data) > remaining {
			return 0, errors.New("error: body bigger than content-length")
		}

		r.Body = append(r.Body, data...)

		if len(r.Body) == cLength {
			r.state = done
			return len(data), nil
		}

		return len(data), nil

	case done:

		return 0, errors.New("error: try to parse in a done state")

	default:
		return 0, errors.New("error: unknown state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, lineSeparator)
	if idx == -1 {
		return nil, 0, nil
	}

	line := data[:idx]
	consumed := idx + len(lineSeparator)

	parts := bytes.Split(line, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, errors.New("invalid request line")
	}

	for _, r := range parts[0] {
		if r < 'A' || r > 'Z' {
			return nil, 0, errors.New("invalid method")
		}
	}

	versionParts := bytes.Split(parts[2], []byte("/"))

	if len(versionParts) != 2 || string(versionParts[0]) != "HTTP" || string(versionParts[1]) != "1.1" {
		return nil, 0, errors.New("invalid http version")
	}

	return &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(versionParts[1]),
	}, consumed, nil
}
