package css

import (
	"strings"
	"unicode"
)

type Parser struct {
	s string
	i int
}

func (p *Parser) whitespace() {
	for p.i < len(p.s) && unicode.IsSpace(rune(p.s[p.i])) {
		p.i++
	}
}

func (p *Parser) word() string {
	start := p.i
	for p.i < len(p.s) {
		r := rune(p.s[p.i])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || contains("#-.%", r) {
			p.i++
		} else {
			break
		}
	}
	if start < p.i {
		panic("unexpected end of word")
	}
	return p.s[start:p.i]
}

func (p *Parser) literal(literal rune) {
	if !(p.i < len(p.s) && rune(p.s[p.i]) == literal) {
		panic("expected string literal")
	}
	p.i++
}

func (p *Parser) pair() (string, string) {
	prop := p.word()
	p.whitespace()
	p.literal(':')
	p.whitespace()
	value := p.word()
	return strings.ToLower(prop), value
}

func (p *Parser) parse() map[string]string {
	pairs := make(map[string]string)
	for p.i < len(p.s) {
		prop, value := p.pair()
		pairs[prop] = value
		p.whitespace()
		p.literal(';')
		p.whitespace()
	}
	return pairs
}

func contains(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
