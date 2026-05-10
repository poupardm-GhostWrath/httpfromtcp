// Package headers
package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	crlf = "\r\n"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		// Not enough data
		return 0, false, nil
	}
	if idx == 0 {
		// Starts with CRLF
		return 2, true, nil
	}

	headerLineText := string(data[:idx])
	headerLine, err := headerFromString(headerLineText)
	if err != nil {
		return 0, false, err
	}
	h[headerLine[0]] = headerLine[1]
	return len(headerLineText) + 2, false, nil
}

func headerFromString(str string) ([]string, error) {
	fieldName, fieldValue, found := strings.Cut(str, ":")
	if !found {
		return nil, fmt.Errorf("poorly formatted header: %s", str)
	}

	// Check if field name is valid
	// Check leading, trailing and in text whitespace
	if strings.Contains(fieldName, " ") || strings.Contains(fieldName, "\t") {
		return nil, fmt.Errorf("poorly formatted field name: %s", str)
	}

	// Trim field value whitespace
	fieldValue = strings.TrimSpace(fieldValue)

	return []string{fieldName, fieldValue}, nil
}
