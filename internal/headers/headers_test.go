package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	consumed, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	value, _ := headers.Get("hoST")
	assert.Equal(t, "localhost:42069", value)
	assert.Equal(t, 23, consumed)
	assert.False(t, done)
	_, done, _ = headers.Parse(data[consumed:])
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("    Host : localhost:42069  \r\n\r\n")
	consumed, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, consumed)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host:                 localhost:42069  \r\n\r\n")
	consumed, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, len(data) - 2, consumed)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Content-type: application/json\r\nHost:localhost:42069\r\n\r\n")
	consumed, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 32, consumed)
	assert.False(t, done)
	consumed, done, err = headers.Parse(data[consumed:])
	require.NoError(t, err)
	assert.Equal(t, 22, consumed)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("ConTent-Type: application/json\r\n\r\n")
	consumed, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, len(data) - 2, consumed)
	assert.False(t, done)
	value,ok := headers.Get("content-type")
	require.True(t, ok)
	assert.Equal(t, "application/json", value)

	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	consumed, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, consumed)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte(": application/json\r\n\r\n")
	consumed, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, consumed)
	assert.False(t, done)


	headers = NewHeaders()
	data = []byte("content-type: application/json\r\ncontent-type:    text/html\r\n")
	consumed, done, err = headers.Parse(data)
	consumed, done, err = headers.Parse(data[consumed:])
	ct, _ := headers.Get("content-type")
	assert.Equal(t, "application/json, text/html", ct)
}

func TestHeadersAllIn(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Content-type: application/json\r\nHost:localhost:42069\r\n\r\n")
	consumed, done, err := headers.ParseAllIn(data)
	require.NoError(t, err)
	assert.Equal(t, len(data), consumed)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("content-type: application/json\r\ncontent-type:    text/html\r\n")
	consumed, done, err = headers.ParseAllIn(data)
	consumed, done, err = headers.ParseAllIn(data[consumed:])
	ct, _ := headers.Get("content-type")
	assert.Equal(t, "application/json, text/html", ct)
}


