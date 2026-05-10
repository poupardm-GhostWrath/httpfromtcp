// Package headers
package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

const (
	crlf = "\r\n"
)

var specialCharacters = []string{"!", "#", "$", "%", "&", "'", "*", "+", "-", ".", "^", "_", "`", "|", "~"}

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
	h.Set(headerLine[0], headerLine[1])
	return len(headerLineText) + 2, false, nil
}

func headerFromString(str string) ([]string, error) {
	fieldName, fieldValue, found := strings.Cut(str, ":")
	if !found {
		return nil, fmt.Errorf("poorly formatted header: %s", str)
	}

	// Check if field name is valid
	// Check if field name is more than 1 character
	if len(fieldName) < 1 {
		return nil, fmt.Errorf("poorly formatted field name: %s", str)
	}
	// Check leading, trailing and in text whitespace
	if strings.Contains(fieldName, " ") || strings.Contains(fieldName, "\t") {
		return nil, fmt.Errorf("poorly formatted field name: %s", str)
	}
	// lowercase field name
	fieldName = strings.ToLower(fieldName)

	// Check for invalid characters
	for _, r := range fieldName {
		isLetter := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isSpecial := slices.Contains(specialCharacters, string(r))
		if !isLetter && !isDigit && !isSpecial {
			return nil, fmt.Errorf("poorly formatted field name: %s", str)
		}
	}

	// Trim field value whitespace
	fieldValue = strings.TrimSpace(fieldValue)

	return []string{fieldName, fieldValue}, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{v, value}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	val, ok := h[key]
	return val, ok
}
