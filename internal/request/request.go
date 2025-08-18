package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mahauni/http_scratch/internal/headers"
)

type parseState string

const (
	StateInit    parseState = "init"
	StateHeaders parseState = "headers"
	StateDone    parseState = "done"
	StateBody    parseState = "Body"
	StateError   parseState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string

	state parseState
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("Malformed request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported http version")
var ERROR_METHOD_NOT_VALID = fmt.Errorf("Method not valid")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("Request in error state")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VERSION
	}

	if string(parts[0]) != strings.ToUpper(string(parts[0])) {
		return nil, 0, ERROR_METHOD_NOT_VALID
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func (r *Request) hasBody() bool {
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				r.state = StateDone
				break
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("This should never occour")
		}
	}

	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
