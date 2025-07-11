// package tenmusu
package main

import (
	"bufio"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
)

type URL struct {
	scheme string
	host   string
	port   int
	path   string
}

type Response struct {
	status      string
	version     string
	explanation string
	headers     map[string]string
	body        string
}

func NewURL(url string) *URL {
	ret := &URL{}

	parts := strings.Split(url, "://")
	if len(parts) != 2 {
		return nil // Invalid URL format
	}
	ret.scheme = parts[0]
	switch ret.scheme {
	case "http":
		ret.port = 80
	case "https":
		ret.port = 443
	}
	part := parts[1]

	if strings.Contains(part, "/") {
		parts = strings.SplitN(part, "/", 2)
		ret.host = parts[0]
		ret.path = "/" + parts[1]
	} else {
		ret.host = part
		ret.path = ""
	}
	if strings.Contains(ret.host, ":") {
		parts = strings.SplitN(ret.host, ":", 2)
		ret.host = parts[0]
		ret.port, _ = strconv.Atoi(parts[1])
	}

	return ret
}

func (u *URL) request() *Response {
	conn, err := net.Dial("tcp", u.host+":"+strconv.Itoa(u.port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if u.scheme == "https" {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName: u.host,
		})
		if err := tlsConn.Handshake(); err != nil {
			panic(err)
		}
		defer tlsConn.Close()
		conn = tlsConn
	}

	request := "GET " + u.path + " HTTP/1.1\r\n"
	request += "Host: " + u.host + "\r\n"
	request += "\r\n"
	conn.Write([]byte(request))

	reader := bufio.NewReader(conn)

	// status line
	statusLine, _ := readLine(reader)
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		println("Invalid response from server")
		return nil
	}
	version := parts[0]
	status := parts[1]
	explanation := parts[2]

	// headers
	responseHeaders := make(map[string]string)
	for {
		line, _ := readLine(reader)
		if line == "" {
			break // End of headers
		}
		parts := strings.SplitN(line, ":", 2)
		responseHeaders[strings.ToLower(strings.TrimSpace(parts[0]))] = strings.TrimSpace(parts[1])
	}

	// body
	size, err := strconv.Atoi(responseHeaders["content-length"])
	if err != nil {
		println("Content-Length header not found or invalid")
		size = 0
	}
	buf := make([]byte, size)

	_, err = reader.Read(buf)
	content := string(buf)

	return &Response{
		status:      status,
		version:     version,
		explanation: explanation,
		headers:     responseHeaders,
		body:        content,
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func main() {
	url := NewURL("https://example.org/")
	if url != nil {
		response := url.request()
		println("\nStatus line:")
		println(response.version + " " + response.status + " " + response.explanation)
		println("\nResponse headers:")
		for key, value := range response.headers {
			println(key + ": " + value)
		}
		println("\nResponse body:")
		println(response.body)
	} else {
		println("Invalid URL")
	}
}
