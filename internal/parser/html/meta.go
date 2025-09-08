package html

import (
	"strings"

	"github.com/pishiko/tenmusu/internal/parser/model"
)

type MetaInfo struct {
	CssLinks []string
	Charset  string
}

func (p *MetaInfo) recursive(node *model.Node) {
	if node.Type == model.Element {
		switch node.Value {
		case "link":
			if rel, ok := node.Attrs["rel"]; ok && rel == "stylesheet" {
				if href, ok := node.Attrs["href"]; ok {
					p.CssLinks = append(p.CssLinks, href)
				}
			}
		case "meta":
			if charset, ok := node.Attrs["charset"]; ok {
				p.Charset = charset
			}
			if httpEquiv, ok := node.Attrs["http-equiv"]; ok {
				if strings.ToLower(httpEquiv) == "content-type" {
					if content, ok := node.Attrs["content"]; ok {
						contentType := content
						parts := splitAndTrim(contentType, ";")
						if len(parts) > 0 {
							if strings.HasPrefix(strings.ToLower(parts[0]), "charset=") {
								parts = splitAndTrim(parts[0], "=")
							}
							if len(parts) > 1 {
								p.Charset = strings.TrimSpace(parts[1])
							}
						}
					}
				}
			}
		}
	}
	for _, child := range node.Children {
		p.recursive(child)
	}
}

func splitAndTrim(contentType, s string) []string {
	parts := strings.Split(contentType, s)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func afterParse(node *model.Node) MetaInfo {
	p := &MetaInfo{
		CssLinks: []string{},
		Charset:  "utf-8",
	}
	p.recursive(node)
	println("Detected charset:", p.Charset)
	return *p
}
