package html

type TokenType int

type Token struct {
	Type     TokenType
	Value    string
	Children []Token
}

const (
	Text TokenType = iota
	Tag
)

type Parser struct {
	body       string
	unfinished []Token
}

func Parse(body string) []Token {
	parser := &Parser{
		body:       body,
		unfinished: []Token{},
	}
	return parser.tokens()
}

func (p *Parser) tokens() []Token {
	isInTag := false
	buf := ""
	tokens := []Token{}
	for _, char := range p.body {
		switch char {
		case '<':
			isInTag = true
			if buf != "" {
				tokens = append(tokens, Token{Type: Text, Value: buf})
				buf = ""
			}
		case '>':
			isInTag = false
			tokens = append(tokens, Token{Type: Tag, Value: buf})
			buf = ""
		default:
			buf += string(char)
		}
	}
	if !isInTag && buf != "" {
		tokens = append(tokens, Token{Type: Text, Value: buf})
	}
	return tokens
}
