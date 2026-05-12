// Package request
package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/poupardm-GhostWrath/httpfromtcp/internal/headers"
)

const (
	crlf       = "\r\n"
	bufferSize = 32
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	state       requestState
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HTTPVersion   string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	r := &Request{state: requestStateInitialized, Headers: headers.NewHeaders()}

	for r.state != requestStateDone {
		// Check Buffer
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// Read into buffer
		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if r.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", r.state, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		// Parse buffer
		numBytesParsed, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[numBytesParsed:])

		readToIndex -= numBytesParsed
	}

	return r, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	// Split string into parts
	parts := strings.Split(str, " ")

	// Check if 3 parts exists
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	// Extract method && Check if Method is CAPITAL letters only
	method := parts[0]
	for _, r := range method {
		if r < 'A' || r > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	// Extract request target
	requestTarget := parts[1]

	// Extract HTTP version && Check if HTTP Version is 1.1
	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}
	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	httpVersion := versionParts[1]
	if httpVersion != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpVersion)
	}

	// Return Request line
	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HTTPVersion:   httpVersion,
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, num, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if num == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return num, nil
	case requestStateParsingHeaders:
		num, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return num, nil
	case requestStateParsingBody:
		val, ok := r.Headers.Get("Content-Length")
		if !ok {
			// No Content-Length Header
			r.state = requestStateDone
			return 0, nil
		}
		n, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > n {
			return 0, fmt.Errorf("error: body is greater than the Content-Length")
		}
		if len(r.Body) == n {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}
