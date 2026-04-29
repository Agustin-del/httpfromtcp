package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		make(map[string]string),
	}
}

func (h *Headers) Get(key string) (string, bool){
	value, ok := h.headers[strings.ToLower(key)]

	return value, ok
}

func (h *Headers) Set(key, value string) {
	lowerKey := strings.ToLower(key)
	trimValue := strings.TrimSpace(value)
	if v, ok := h.headers[lowerKey]; ok{
		h.headers[lowerKey] = v + fmt.Sprintf(", %s", trimValue)
		return
	} 

	h.headers[lowerKey] = trimValue
}

var lineSeparator = []byte("\r\n")


func (h Headers) Parse(data []byte) (n int, done bool, err error) {
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

	isValid := validToken(key)
	if !isValid {
		return 0, false, errors.New("invalid field name")
	}

	h.Set(string(key), string(header[1]))
	return len(line) + len(lineSeparator), false, nil
}

func validToken(key []byte) bool {
	if len(key) == 0 {
		return false
	}

	for _, c := range key {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '!' || c == '#' || c == '$' || c == '%' ||
			c == '&' || c == '\'' || c == '*' || c == '+' ||
			c == '-' || c == '.' || c == '^' || c == '_' ||
			c == '`' || c == '|' || c == '~') {
			return false
		}
	}

	return true

}
