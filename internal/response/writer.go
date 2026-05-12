package response

import (
	"fmt"
	"io"

	"github.com/poupardm-GhostWrath/httpfromtcp/internal/headers"
)

const crlf = "\r\n"

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
)

type Writer struct {
	writerState writerState
	writer      io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: writerStateStatusLine,
		writer:      w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("error writer not in correct state: %d", w.writerState)
	}
	defer func() { w.writerState = writerStateHeaders }()
	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("error writer not in correct state: %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()
	for key, value := range h {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("error writer not in correct state: %d", w.writerState)
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("error writer not in correct state: %d", w.writerState)
	}

	totalBytes := 0
	n, err := w.writer.Write(fmt.Appendf([]byte{}, "%x%s", len(p), crlf))
	if err != nil {
		return totalBytes, err
	}
	totalBytes += n
	n, err = w.writer.Write(p)
	if err != nil {
		return totalBytes, err
	}
	totalBytes += n
	n, err = w.writer.Write([]byte(crlf))
	if err != nil {
		return totalBytes, err
	}
	totalBytes += n
	return totalBytes, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("error writer not in correct state: %d", w.writerState)
	}
	n, err := w.writer.Write(fmt.Appendf([]byte{}, "0%s", crlf))
	if err != nil {
		return n, err
	}
	w.writerState = writerStateTrailers
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		return fmt.Errorf("error writer not in correct state: %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()
	for key, value := range h {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte(crlf))
	return err
}
