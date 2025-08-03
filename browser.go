package main

import (
	"os"
	"sort"

	"github.com/pishiko/tenmusu/internal/http"
	"github.com/pishiko/tenmusu/internal/parser/css"
	"github.com/pishiko/tenmusu/internal/parser/html"
	"github.com/pishiko/tenmusu/internal/parser/model"
	"github.com/pishiko/tenmusu/internal/window"
)

type Browser struct {
}

func NewBrowser() *Browser {
	return &Browser{}
}
func (b *Browser) Load(url string) {
	docUrl := http.NewURL(url)
	if docUrl == nil {
		println("Invalid URL")
		return
	}

	response := docUrl.Request()
	println("\nStatus line:")
	println(response.Version + " " + response.Status + " " + response.Explanation)
	println("\nResponse headers:")
	for key, value := range response.Headers {
		println(key + ": " + value)
	}

	node := html.Parse(response.Body)

	// css
	cssLinks := afterParse(node)
	// browser.css
	cssContent, err := os.ReadFile("browser.css")
	if err != nil {
		println("Error reading browser.css:", err)
		return
	}
	rules := css.CSSParse(string(cssContent))
	// cssLinks
	for _, link := range cssLinks {
		cssUrl := docUrl.Resolve(link)
		println("Fetching CSS from:", cssUrl.Scheme+"://"+cssUrl.Host+cssUrl.Path)
		println("link:", link)
		response := cssUrl.Request()
		if response == nil {
			println("Failed to fetch CSS from:", link)
			continue
		}
		rules = append(rules, css.CSSParse(response.Body)...)
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Selector.Priority() > rules[j].Selector.Priority()
	})
	css.ApplyStyle(&node, rules)

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

func printDebug(n model.Node, indent int) {
	for i := 0; i < indent; i++ {
		print("  ")
	}
	switch n.Type {
	case model.Text:
		println(n.Value)
	case model.Element:
		println("<" + n.Value + ">")
		for _, node := range n.Children {
			printDebug(node, indent+1)
		}
		for i := 0; i < indent; i++ {
			print("  ")
		}
		println("</" + n.Value + ">")
	}
}

type AfterParser struct {
	cssLinks []string
}

func (p *AfterParser) recursive(n model.Node) {
	if n.Type == model.Element && n.Value == "link" {
		if rel, ok := n.Attrs["rel"]; ok && rel == "stylesheet" {
			if href, ok := n.Attrs["href"]; ok {
				p.cssLinks = append(p.cssLinks, href)
			}
		}
	}
	for _, child := range n.Children {
		p.recursive(child)
	}
}

func afterParse(n model.Node) []string {
	p := &AfterParser{
		cssLinks: []string{},
	}
	p.recursive(n)
	return p.cssLinks
}
