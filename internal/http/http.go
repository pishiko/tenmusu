package http

import (
	"bufio"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
)

type URL struct {
	Scheme string
	Host   string
	Port   int
	Path   string
}

type Response struct {
	Status      string
	Version     string
	Explanation string
	Headers     map[string]string
	Body        string
}

func NewURL(url string) *URL {
	ret := &URL{}

	parts := strings.Split(url, "://")
	if len(parts) != 2 {
		return nil // Invalid URL format
	}
	ret.Scheme = parts[0]
	switch ret.Scheme {
	case "http":
		ret.Port = 80
	case "https":
		ret.Port = 443
	}
	part := parts[1]

	if strings.Contains(part, "/") {
		parts = strings.SplitN(part, "/", 2)
		ret.Host = parts[0]
		ret.Path = "/" + parts[1]
	} else {
		ret.Host = part
		ret.Path = ""
	}
	if strings.Contains(ret.Host, ":") {
		parts = strings.SplitN(ret.Host, ":", 2)
		ret.Host = parts[0]
		ret.Port, _ = strconv.Atoi(parts[1])
	}

	return ret
}

func (u *URL) Request() *Response {
	if u.Scheme == "file" {
		return u.openFile()
	}

	conn, err := net.Dial("tcp", u.Host+":"+strconv.Itoa(u.Port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if u.Scheme == "https" {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName: u.Host,
		})
		if err := tlsConn.Handshake(); err != nil {
			panic(err)
		}
		defer tlsConn.Close()
		conn = tlsConn
	}

	request := "GET " + u.Path + " HTTP/1.1\r\n"
	request += "Host: " + u.Host + "\r\n"
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

	for offset := 0; size > offset; {
		n, err := reader.Read(buf[offset:])
		if err != nil {
			println("Error reading response body:", err.Error())
			return nil
		}
		offset += n
	}

	content := string(buf)

	return &Response{
		Status:      status,
		Version:     version,
		Explanation: explanation,
		Headers:     responseHeaders,
		Body:        content,
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
