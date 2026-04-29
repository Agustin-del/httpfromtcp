package headers

import (
	"bytes"
	"errors"
)

type Headers map[string]string

func NewHeaders() Headers{
	return make(Headers)
}

var lineSeparator = []byte("\r\n")

func (h Headers) Parse(data []byte) (n int, done bool, err error){
	//capaz hay que pasarlo a que parsee todos los headers porque quien lo hizo es un dolobu jaja
	endLine := bytes.Index(data, lineSeparator)
	if endLine == -1 {
		return 0, false, nil
	}

	if endLine == 0 {
		return len(lineSeparator), true, nil
	}

	line := data[:endLine]

	header := bytes.SplitN(line, []byte(":"), 2)
	if len(header) != 2 {
		return 0, false, errors.New("invalid header missing colon")
	}

	rawKey := header[0]
	key := bytes.TrimSpace(rawKey)

	if !bytes.Equal(rawKey, key) {
		return 0, false, errors.New("invalid spacing in field name")
	}

	if bytes.ContainsAny(key, " \t") {
		return 0, false, errors.New("invalid field name")
	}
	
	h[string(key)] = string(bytes.TrimSpace(header[1]))
	return len(line) + len(lineSeparator), false, nil
}
