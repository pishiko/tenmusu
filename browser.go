package main

import (
	"github.com/pishiko/tenmusu/internal/html"
	"github.com/pishiko/tenmusu/internal/http"
	"github.com/pishiko/tenmusu/internal/window"
)

type Browser struct {
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

	tokens := html.Lex(response.Body)
	window.Open(tokens)
}

func main() {
	browser := NewBrowser()
	browser.Load("https://example.org/")
}
