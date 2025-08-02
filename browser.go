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

	nodes := html.Parse(response.Body)
	printDebug(nodes, 0)
	window.Open(nodes)
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

func printDebug(nodes []html.Node, indent int) {
	for _, node := range nodes {
		for i := 0; i < indent; i++ {
			print("  ")
		}
		switch node.Type {
		case html.Text:
			println(node.Value)
		case html.Element:
			println("<" + node.Value + ">")
			printDebug(node.Children, indent+1)
			for i := 0; i < indent; i++ {
				print("  ")
			}
			println("</" + node.Value + ">")
		}
	}
}
