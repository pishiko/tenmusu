package css

import (
	"errors"
	"strings"
	"unicode"

	"github.com/pishiko/tenmusu/internal/parser/model"
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

func (p *CSSParser) ignoreUntil(s string) rune {
	for p.i < len(p.s) {
		if contains(s, rune(p.s[p.i])) {
			return rune(p.s[p.i])
		}
		p.i++
	}
	return 0 // End of string
}

func (p *CSSParser) skipSpaceAndComments() {
	p.whitespace()
	for ok := p.comment(); ok; ok = p.comment() {
		p.whitespace()
	}
}

func (p *CSSParser) whitespace() {
	for p.i < len(p.s) && unicode.IsSpace(rune(p.s[p.i])) {
		p.i++
	}
}

func (p *CSSParser) comment() bool {
	if p.i+1 < len(p.s) && p.s[p.i] == '/' && p.s[p.i+1] == '*' {
		p.i += 2 // skip /*
		for p.i+1 < len(p.s) && !(p.s[p.i] == '*' && p.s[p.i+1] == '/') {
			p.i++
		}
		if p.i+1 < len(p.s) {
			p.i += 2 // skip */
			return true
		} else {
			println("[CSS PARSER] unterminated comment")
		}
	}
	return false
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
		println("[CSS PARSER] unexpected end of word", p.i)
		// println(p.s, " at ", p.i)
		return ""
	}
	return p.s[start:p.i]
}

func (p *CSSParser) literal(literal rune) {
	if !(p.i < len(p.s) && rune(p.s[p.i]) == literal) {
		print("[CSS PARSER] expected string literal ", string(literal))
		if p.i < len(p.s) {
			println(" but got ", string(p.s[p.i]))
		} else {
			println(" but got end of string")
		}
	}
	p.i++
}

func (p *CSSParser) pair() (string, string, error) {
	prop := p.word()
	if prop == "" {
		return "", "", errors.New("[CSS PARSER] expected property name")
	}
	p.skipSpaceAndComments()
	p.literal(':')
	p.skipSpaceAndComments()
	value := p.word()
	if value == "" {
		return "", "", errors.New("[CSS PARSER] expected property value")
	}
	return strings.ToLower(prop), value, nil
}

func (p *CSSParser) body() map[string]string {
	pairs := make(map[string]string)
	for p.i < len(p.s) && rune(p.s[p.i]) != '}' {
		prop, value, err := p.pair()
		if err != nil {
			println("[CSS PARSER] Error parsing property:", err.Error())
			why := p.ignoreUntil(";}")
			if why == ';' {
				p.literal(';')
			}
			continue
		}
		pairs[prop] = value
		p.skipSpaceAndComments()
		p.literal(';')
		p.skipSpaceAndComments()
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
		p.skipSpaceAndComments()
		selector, err := p.selector()
		if err != nil {
			println("[CSS PARSER] Error parsing selector:", err.Error())
			why := p.ignoreUntil("}")
			if why == '}' {
				p.literal('}')
			}
			continue
		}
		p.literal('{')
		p.skipSpaceAndComments()
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

func (p *CSSParser) selector() (Selector, error) {
	ret := Selector(nil)
	tag := p.word()
	if tag == "" {
		return nil, errors.New("[CSS PARSER] expected tag name")
	}
	ret = &TagSelector{tag: strings.ToLower(tag)}
	p.skipSpaceAndComments()
	for p.i < len(p.s) && rune(p.s[p.i]) != '{' {
		tag := p.word()
		if tag == "" {
			return nil, errors.New("[CSS PARSER] expected tag name")
		}
		decsendant := TagSelector{tag: strings.ToLower(tag)}
		ret = &DescendantSelector{ancestor: ret, descendant: &decsendant}
		p.skipSpaceAndComments()
	}
	return ret, nil
}

func contains(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func ApplyStyle(n *model.Node, rules []CSSRule) {
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
		ApplyStyle(&n.Children[i], rules)
	}
}
