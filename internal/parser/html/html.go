package html

import (
	"strings"

	"github.com/pishiko/tenmusu/internal/parser/model"
	"github.com/pishiko/tenmusu/internal/util"
)

type Parser struct {
	body       string
	unfinished util.Stack[*model.Node]
	node       *model.Node
}

func Parse(body string) *model.Node {
	parser := &Parser{
		body:       body,
		unfinished: util.Stack[*model.Node]{},
	}
	parser.parse()
	return parser.node
}

func (p *Parser) parse() {
	isInTag := false
	buf := ""
	for _, char := range p.body {
		switch char {
		case '<':
			if isInTag {
				buf += string(char)
				continue
			}
			isInTag = true
			if buf != "" {
				p.addText(buf)
				buf = ""
			}
		case '>':
			if isInTag && strings.HasPrefix(buf, "!--") {
				if strings.HasSuffix(buf, "--") {
					isInTag = false
					buf = ""
					continue
				} else {
					buf += string(char)
					continue
				}
			}
			isInTag = false
			p.addElement(buf)
			buf = ""
		default:
			buf += string(char)
		}
	}
	if !isInTag && buf != "" {
		p.addText(buf)
	}

	// finish
	for node := p.unfinished.Pop(); node != nil; node = p.unfinished.Pop() {
		p.node = node
	}
}

func (p *Parser) addElement(text string) {
	if text == "" {
		return
	}
	parts := strings.Split(strings.ReplaceAll(text, "\n", " "), " ")
	name := parts[0]

	// attr TODO FIX
	attrs := make(map[string]string)
	for _, part := range parts[1:] {
		if strings.Contains(part, "=") {
			attrParts := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(attrParts[0])
			value := strings.TrimSpace(attrParts[1])
			if len(value) > 1 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			attrs[key] = value
		}
	}

	if name[0] == '/' {
		name = name[1:]
		if node := p.unfinished.Pop(); node != nil {
			if node.Value != name {
				println("Mismatched closing tag: " + name + " for " + node.Value)
			}
			if parent := p.unfinished.Peek(); parent != nil {
				parent.Children = append(parent.Children, node)
			} else {
				p.node = node
			}
		} else {
			panic("Unmatched closing tag: " + name)
		}
	} else {
		if isSelefClosingTag(name) {
			if parent := p.unfinished.Peek(); parent != nil {
				parent.Children = append(parent.Children, &model.Node{Type: model.Element, Value: name, Parent: parent, Attrs: attrs})
			} else {
				p.node = &model.Node{Type: model.Element, Value: name, Attrs: attrs}
			}
			return
		}
		if parent := p.unfinished.Peek(); parent != nil {
			node := &model.Node{Type: model.Element, Value: name, Parent: parent, Attrs: attrs}
			p.unfinished.Push(node)
		} else {
			node := &model.Node{Type: model.Element, Value: name, Parent: nil, Attrs: attrs}
			p.unfinished.Push(node)
		}
	}
}

func (p *Parser) addText(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	text = replaceCharReference(text)
	if parent := p.unfinished.Peek(); parent != nil {
		parent.Children = append(parent.Children, &model.Node{Type: model.Text, Value: text, Parent: parent})
	} else {
		p.node = &model.Node{Type: model.Text, Value: text}
	}
}

func replaceCharReference(text string) string {
	var characterReferences = map[string]string{
		"&lt;":   "<",
		"&gt;":   ">",
		"&amp;":  "&",
		"&quot;": "\"",
		"&apos;": "'",
	}
	for k, v := range characterReferences {
		text = strings.ReplaceAll(text, k, v)
	}
	return text
}

func isSelefClosingTag(tag string) bool {
	var SELF_CLOSING_TAGS = [...]string{
		"area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr",
	}
	for _, selfClosingTag := range SELF_CLOSING_TAGS {
		if tag == selfClosingTag {
			return true
		}
	}
	return false
}
