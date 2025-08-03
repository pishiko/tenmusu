package html

import (
	"strings"
	"unicode"
)

type CSSParser struct {
	s string
	i int
}

func InlineCSSParse(s string) map[string]string {
	parser := &CSSParser{
		s: s,
		i: 0,
	}
	return parser.body()
}

func CSSParse(s string) []CSSRule {
	parser := &CSSParser{
		s: s,
		i: 0,
	}
	return parser.parse()
}

func (p *CSSParser) whitespace() {
	for p.i < len(p.s) && unicode.IsSpace(rune(p.s[p.i])) {
		p.i++
	}
}

func (p *CSSParser) word() string {
	start := p.i
	for p.i < len(p.s) {
		r := rune(p.s[p.i])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || contains("#-.%", r) {
			p.i++
		} else {
			break
		}
	}
	if start >= p.i {
		println("[WARNING][CSS PARSER] unexpected end of word")
		println(p.s, " at ", p.i)
	}
	return p.s[start:p.i]
}

func (p *CSSParser) literal(literal rune) {
	if !(p.i < len(p.s) && rune(p.s[p.i]) == literal) {
		println("[WARNING][CSS PARSER] expected string literal")
		print("   "+p.s, " at ", p.i, " expected ", string(literal))
		if p.i < len(p.s) {
			println(" but got ", string(p.s[p.i]))
		} else {
			println(" but got end of string")
		}
	}
	p.i++
}

func (p *CSSParser) pair() (string, string) {
	prop := p.word()
	p.whitespace()
	p.literal(':')
	p.whitespace()
	value := p.word()
	return strings.ToLower(prop), value
}

func (p *CSSParser) body() map[string]string {
	pairs := make(map[string]string)
	for p.i < len(p.s) && rune(p.s[p.i]) != '}' {
		prop, value := p.pair()
		pairs[prop] = value
		p.whitespace()
		p.literal(';')
		p.whitespace()
	}
	return pairs
}

type CSSRule struct {
	Selector Selector
	Body     map[string]string
}

func (p *CSSParser) parse() []CSSRule {
	rules := []CSSRule{}
	for p.i < len(p.s) {
		p.whitespace()
		selector := p.selector()
		p.literal('{')
		p.whitespace()
		body := p.body()
		p.literal('}')
		rule := CSSRule{
			Selector: selector,
			Body:     body,
		}
		rules = append(rules, rule)
	}
	return rules
}

func (p *CSSParser) selector() Selector {
	ret := Selector(nil)
	ret = &TagSelector{tag: strings.ToLower(p.word())}
	p.whitespace()
	for p.i < len(p.s) && rune(p.s[p.i]) != '{' {
		tag := p.word()
		decsendant := TagSelector{tag: strings.ToLower(tag)}
		ret = &DescendantSelector{ancestor: ret, descendant: &decsendant}
		p.whitespace()
	}
	return ret
}

func contains(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

type Selector interface {
	Matches(node *Node) bool
}

type TagSelector struct {
	tag string
}

func (ts *TagSelector) Matches(node *Node) bool {
	return node.Type == Element && node.Value == ts.tag
}

type DescendantSelector struct {
	ancestor   Selector
	descendant Selector
}

func (ds *DescendantSelector) Matches(node *Node) bool {
	if !ds.ancestor.Matches(node) {
		return false
	}
	for node.Parent != nil {
		if ds.ancestor.Matches(node.Parent) {
			return true
		}
		node = node.Parent
	}
	return false
}
