package request

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}

	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))

	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {

	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: len("GET / HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n"),
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	reader = &chunkReader{
		data:            "POST /coffee HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 8,
	}

	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	reader = &chunkReader{
		data:            "get /fog HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 4,
	}

	_, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 8,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data:            "/coffee GET HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 10,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data:            "/coffee GET HTTP/2.0\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 6,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestHeaders(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: len("GET / HTTP/1.1\r\nHost:localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n"),
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	value, _ := r.Headers.Get("host")
	assert.Equal(t, "localhost:42069", value)
	value, _ = r.Headers.Get("user-agent")
	assert.Equal(t, "curl/8.19.0", value)
	value, _ = r.Headers.Get("accept")
	assert.Equal(t, "*/*", value)

	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestBody(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n", numBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}

	_, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}

	r, err = RequestFromReader(reader)
	require.NoError(t, err)

	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"content",
		numBytesPerRead: 3,
	}

	_, err = RequestFromReader(reader)
	require.NoError(t, err)
}

func TestFail(t *testing.T) {
	fmt.Println("hola")

	reader := &chunkReader{
		data: "GET /hola HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/8.20.0\r\n" +
			"Accept: */*\r\n\r\n",
		numBytesPerRead: 8,
	}

	req, err := RequestFromReader(reader)
	require.NoError(t, err)

	assert.Equal(t, req.RequestLine.RequestTarget, "/hola")

}
