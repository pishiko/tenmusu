package main

import (
	"os"

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

	node := html.Parse(response.Body)
	// browser.css
	cssContent, err := os.ReadFile("browser.css")
	if err != nil {
		println("Error reading browser.css:", err)
		return
	}
	rules := html.CSSParse(string(cssContent))
	node.ApplyStyle(rules)

	// printDebug(node, 0)
	window.Open(node)
}

func main() {
	browser := NewBrowser()
	// 第一引数をURLとして受け取る
	if len(os.Args) < 2 {
		println("Usage: tenmusu <url>")
		return
	}
	url := os.Args[1]
	browser.Load(url)
}

func printDebug(node html.Node, indent int) {
	for i := 0; i < indent; i++ {
		print("  ")
	}
	switch node.Type {
	case html.Text:
		println(node.Value)
	case html.Element:
		println("<" + node.Value + ">")
		for _, node := range node.Children {
			printDebug(node, indent+1)
		}
		for i := 0; i < indent; i++ {
			print("  ")
		}
		println("</" + node.Value + ">")
	}
}
