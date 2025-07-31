package html

import (
	"strings"

	"github.com/pishiko/tenmusu/internal/util"
)

type TokenType int

type Token struct {
	Type     TokenType
	Value    string
	Children []Token
	Parent   *Token
}

const (
	Text TokenType = iota
	Element
)

type Parser struct {
	body       string
	unfinished util.Stack[Token]
	tokens     []Token
}

func Parse(body string) []Token {
	parser := &Parser{
		body:       body,
		unfinished: util.Stack[Token]{},
		tokens:     []Token{},
	}
	return parser.parse()
}

func (p *Parser) parse() []Token {
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
		p.tokens = append(p.tokens, node)
	}
	return p.tokens
}

func (p *Parser) addElement(text string) {
	if text == "" {
		return
	}
	name := strings.Split(text, " ")[0] // TODO FIX

	if name[0] == '/' {
		name = name[1:]
		if node, ok := p.unfinished.Pop(); ok {
			if node.Value != name {
				println("Mismatched closing tag: " + name + " for " + node.Value)
			}
			if parent := p.unfinished.Peek(); parent != nil {
				parent.Children = append(parent.Children, node)
			} else {
				p.tokens = append(p.tokens, node)
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
				parent.Children = append(parent.Children, Token{Type: Element, Value: name, Parent: parent})
			} else {
				p.tokens = append(p.tokens, Token{Type: Element, Value: name})
			}
			return
		}
		if parent := p.unfinished.Peek(); parent != nil {
			node := Token{Type: Element, Value: name, Parent: parent}
			p.unfinished.Push(node)
		} else {
			node := Token{Type: Element, Value: name}
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
		parent.Children = append(parent.Children, Token{Type: Text, Value: text, Parent: parent})
	} else {
		p.tokens = append(p.tokens, Token{Type: Text, Value: text})
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
