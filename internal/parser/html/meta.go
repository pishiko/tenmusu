package html

import (
	"strings"

	"github.com/pishiko/tenmusu/internal/parser/model"
)

type AfterParser struct {
	node *model.Node
	meta model.MetaInfo
}

func (p *AfterParser) recursive(node *model.Node) {
	if node.Type == model.Element {
		switch node.Value {
		case "link":
			if rel, ok := node.Attrs["rel"]; ok && rel == "stylesheet" {
				if href, ok := node.Attrs["href"]; ok {
					p.meta.CssLinks = append(p.meta.CssLinks, href)
				}
			}
		case "meta":
			if charset, ok := node.Attrs["charset"]; ok {
				p.meta.Charset = charset
			}
			if httpEquiv, ok := node.Attrs["http-equiv"]; ok {
				if strings.ToLower(httpEquiv) == "content-type" {
					if content, ok := node.Attrs["content"]; ok {
						contentType := content
						ctParts := splitAndTrim(contentType, ";")
						if len(ctParts) > 0 {
							for _, part := range ctParts {
								if strings.HasPrefix(strings.ToLower(part), "charset=") {
									csParts := splitAndTrim(part, "=")
									if len(csParts) > 1 {
										p.meta.Charset = strings.TrimSpace(csParts[1])
									}
								}
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

func afterParse(node *model.Node) model.MetaInfo {
	p := &AfterParser{
		node: node,
		meta: model.MetaInfo{
			CssLinks: []string{},
			Charset:  "utf-8",
		},
	}
	p.recursive(node)
	println("Detected charset:", p.meta.Charset)
	return p.meta
}
