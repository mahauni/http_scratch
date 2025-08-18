package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

type Headers struct {
	headers map[string]string
}

var CRLF = []byte("\r\n")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Malformed field line")
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("Malformed field name")
	}

	return string(name), string(value), nil

}

func (h *Headers) Get(name string) (string, bool) {
	v, ok := h.headers[strings.ToLower(name)]
	return v, ok
}

func (h *Headers) Replace(name, value string) {
	name = strings.ToLower(name)

	h.headers[name] = value
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) Foreach(cb func(n, v string)) {
	for n, v := range h.headers {
		cb(n, v)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {

	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], CRLF)
		if idx == -1 {
			break
		}

		// Empty header
		if idx == 0 {
			done = true
			read += len(CRLF)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("Malformed header name")
		}

		read += idx + len(CRLF)
		h.Set(name, value)

	}

	return read, done, nil
}
