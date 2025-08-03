package html

import (
	"strings"

	"github.com/pishiko/tenmusu/internal/util"
)

type NodeType int

const (
	Text NodeType = iota
	Element
)

type Node struct {
	Type     NodeType
	Value    string
	Children []Node
	Parent   *Node
	Attrs    map[string]string
	Style    map[string]string
}

func (n *Node) ApplyStyle(rules []CSSRule) {
	if n.Style == nil {
		n.Style = make(map[string]string)
	}
	// sheet
	for _, rule := range rules {
		if !rule.Selector.Matches(n) {
			continue
		}
		for property, value := range rule.Body {
			n.Style[property] = value
		}
	}
	// inline
	if styleText, ok := n.Attrs["style"]; ok {
		n.Style = InlineCSSParse(styleText)
	}
	// children
	for i := range n.Children {
		n.Children[i].ApplyStyle(rules)
	}
}

type Parser struct {
	body       string
	unfinished util.Stack[Node]
	node       Node
}

func Parse(body string) Node {
	parser := &Parser{
		body:       body,
		unfinished: util.Stack[Node]{},
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
			isInTag = true
			if buf != "" {
				p.addText(buf)
				buf = ""
			}
		case '>':
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
	for node, ok := p.unfinished.Pop(); ok; node, ok = p.unfinished.Pop() {
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
		if node, ok := p.unfinished.Pop(); ok {
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
		if name[0] == '!' {
			return
		}
		if isSelefClosingTag(name) {
			if parent := p.unfinished.Peek(); parent != nil {
				parent.Children = append(parent.Children, Node{Type: Element, Value: name, Parent: parent, Attrs: attrs})
			} else {
				p.node = Node{Type: Element, Value: name, Attrs: attrs}
			}
			return
		}
		if parent := p.unfinished.Peek(); parent != nil {
			node := Node{Type: Element, Value: name, Parent: parent, Attrs: attrs}
			p.unfinished.Push(node)
		} else {
			node := Node{Type: Element, Value: name, Parent: nil, Attrs: attrs}
			p.unfinished.Push(node)
		}
	}
}

func (p *Parser) addText(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if parent := p.unfinished.Peek(); parent != nil {
		parent.Children = append(parent.Children, Node{Type: Text, Value: text, Parent: parent})
	} else {
		p.node = Node{Type: Text, Value: text}
	}
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
