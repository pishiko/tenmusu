package http

import (
	"io"
	"os"
)

func (u *URL) openFile() *Response {
	// open local file
	file, err := os.Open(u.Path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return &Response{
		Status:      "",
		Version:     "",
		Explanation: "",
		Headers:     map[string]string{},
		Body:        string(content),
	}
}
