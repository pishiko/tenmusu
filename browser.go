package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/pishiko/tenmusu/internal/http"
	"github.com/pishiko/tenmusu/internal/parser/css"
	"github.com/pishiko/tenmusu/internal/parser/html"
	"github.com/pishiko/tenmusu/internal/parser/model"
	"github.com/pishiko/tenmusu/internal/window"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Browser struct {
	config struct {
		printNode   bool
		printLayout bool
	}
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

	node, meta := html.Parse(response.Body)
	if meta.Charset != "" {
		if strings.ToLower(meta.Charset) == "shift_jis" {
			reader := transform.NewReader(bytes.NewReader([]byte(response.Body)), japanese.ShiftJIS.NewDecoder())

			body, err := io.ReadAll(reader)
			if err != nil {
				println("Error decoding Shift_JIS:", err)
				return
			}
			node, meta = html.Parse(string(body))
		}
	}

	// browser.css
	cssContent, err := os.ReadFile("browser.css")
	if err != nil {
		println("Error reading browser.css:", err)
		return
	}
	rules := css.CSSParse(string(cssContent))
	// cssLinks
	for _, link := range meta.CssLinks {
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

	if b.config.printNode {
		printDebug(node, 0)
	}
	window.Open(node, b.config.printLayout)
}

func main() {
	nodePrint := flag.Bool("node", false, "Print the parsed HTML node structure")
	layoutPrint := flag.Bool("layout", false, "Print the layout structure")
	flag.Parse()

	flag.Usage = func() {
		fmt.Println("Usage: tenmusu <url> [options] [args...]")
		flag.PrintDefaults()
	}
	// 第一引数をURLとして受け取る
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return
	}
	browser := NewBrowser()
	browser.config.printNode = *nodePrint
	browser.config.printLayout = *layoutPrint
	url := args[0]
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
