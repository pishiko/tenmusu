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
	css.ApplyStyle(node, rules)

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

func printDebug(node *model.Node, indent int) {
	for i := 0; i < indent; i++ {
		print("  ")
	}
	switch node.Type {
	case model.Text:
		println(node.Value)
	case model.Element:
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

type AfterParser struct {
	cssLinks []string
}

func (p *AfterParser) recursive(node *model.Node) {
	if node.Type == model.Element && node.Value == "link" {
		if rel, ok := node.Attrs["rel"]; ok && rel == "stylesheet" {
			if href, ok := node.Attrs["href"]; ok {
				p.cssLinks = append(p.cssLinks, href)
			}
		}
	}
	for _, child := range node.Children {
		p.recursive(child)
	}
}

func afterParse(node *model.Node) []string {
	p := &AfterParser{
		cssLinks: []string{},
	}
	p.recursive(node)
	return p.cssLinks
}
