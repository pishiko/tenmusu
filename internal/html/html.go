package html

type TokenType int

type Token struct {
	Type  TokenType
	Value string
}

const (
	Text TokenType = iota
	Tag
)

func Lex(body string) []Token {
	isInTag := false
	buf := ""
	tokens := []Token{}
	for _, char := range body {
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
