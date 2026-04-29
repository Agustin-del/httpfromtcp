package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	value, _ := headers.Get("hoST")
	assert.Equal(t, "localhost:42069", value)
	assert.Equal(t, 23, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("    Host : localhost:42069  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host:                 localhost:42069  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, len(data) - len("\r\n"), n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Content-type: application/json\r\nHost:localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	n2, done, err := headers.Parse(data[n:])
	n3, done, err := headers.Parse(data[n + n2:])
	require.NoError(t, err)
	assert.Equal(t, len(data), n + n2 + n3)
	assert.True(t, done)


	headers = NewHeaders()
	data = []byte("ConTent-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	n2, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, len(data), n + n2)
	assert.True(t, done)

	value,ok := headers.Get("content-type")
	require.True(t, ok)
	assert.Equal(t, "application/json", value)
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte(": application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)


	headers = NewHeaders()
	data = []byte("content-type: application/json\r\ncontent-type:    text/html\r\n")
	n, done, err = headers.Parse(data)
	n2, done, err = headers.Parse(data[n:])
	ct, _ := headers.Get("content-type")
	assert.Equal(t, "application/json, text/html", ct)
}
