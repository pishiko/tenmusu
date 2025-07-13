package main

import (
	"github.com/pishiko/tenmusu/internal/http"
	"github.com/pishiko/tenmusu/internal/window"
)

type Browser struct {
}

func lex(body string) string {
	isInTag := false
	text := ""
	for _, char := range body {
		if char == '<' {
			isInTag = true
		} else if char == '>' {
			isInTag = false
		} else if !isInTag {
			text += string(char)
		}
	}
	return text
}

func NewBrowser() *Browser {
	return &Browser{}
}
func (b *Browser) Load(url string) {
	u := http.NewURL(url)
	if u == nil {
		println("Invalid URL")
		return
	}

	response := u.Request()
	println("\nStatus line:")
	println(response.Version + " " + response.Status + " " + response.Explanation)
	println("\nResponse headers:")
	for key, value := range response.Headers {
		println(key + ": " + value)
	}

	content := lex(response.Body)
	window.Open(content)
}

func main() {
	browser := NewBrowser()
	browser.Load("https://example.org/")
}
