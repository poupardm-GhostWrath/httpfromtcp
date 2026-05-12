// httpserver
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/poupardm-GhostWrath/httpfromtcp/internal/headers"
	"github.com/poupardm-GhostWrath/httpfromtcp/internal/request"
	"github.com/poupardm-GhostWrath/httpfromtcp/internal/response"
	"github.com/poupardm-GhostWrath/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/video" {
		handlerVideo(w, req)
		return
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handlerProxy(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}

	handler200(w, req)
}

func handler400(w *response.Writer, _ *request.Request) {
	body := `
<html>
	<head>
		<title>400 Bad Request</title>
	</head>
	<body>
		<h1>Bad Request</h1>
		<p>Your request honestly kinda sucked.</p>
	</body>
</html>`
	w.WriteStatusLine(response.StatusBadRequest)
	h := response.GetDefaultHeaders(len([]byte(body)))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}

func handler500(w *response.Writer, _ *request.Request) {
	body := `
<html>
	<head>
		<title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>Okay, you know what? This one is on me.</p>
	</body>
</html>`
	w.WriteStatusLine(response.StatusInternalServerError)
	h := response.GetDefaultHeaders(len([]byte(body)))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}

func handler200(w *response.Writer, _ *request.Request) {
	body := `
<html>
	<head>
		<title>200 OK</title>
	</head>
	<body>
		<h1>Success!</h1>
		<p>Your request was an absolute banger.</p>
	</body>
</html>`
	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(len([]byte(body)))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}

func handlerProxy(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	h.Delete("Content-Length")
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	totalBytes := 0
	buf := make([]byte, maxChunkSize)
	fullBuf := []byte{}
	for {
		n, err := resp.Body.Read(buf)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			fullBuf = append(fullBuf, buf[:n]...)
		}
		totalBytes += n
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}

	hT := headers.NewHeaders()
	hT.Set("X-Content-Length", fmt.Sprintf("%d", totalBytes))
	checkSum := sha256.Sum256(fullBuf)
	hT.Set("X-Content-SHA256", fmt.Sprintf("%x", checkSum))

	err = w.WriteTrailers(hT)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
}

func handlerVideo(w *response.Writer, req *request.Request) {
	data, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		handler500(w, req)
		return
	}
	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(len(data))
	h.Override("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(data)
}
