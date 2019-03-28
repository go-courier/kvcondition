package labelql

import (
	"errors"
	"strconv"
	"strings"
	"text/scanner"
)

var Keywords = map[string]bool{
	"!":   true,
	"=":   true,
	"(":   true,
	")":   true,
	"AND": true,
	"OR":  true,
}

func NewToken(pos scanner.Position, tokenText string) *Token {
	keyword := strings.ToUpper(tokenText)
	isKeyword := Keywords[keyword]

	if isKeyword {
		tokenText = keyword
	}

	return &Token{
		TokenText: tokenText,
		Pos:       pos,
		IsKeyword: isKeyword,
	}
}

type Token struct {
	TokenText string
	Pos       scanner.Position
	IsKeyword bool
}

type TokenScanner struct {
	tokens   []*Token
	priority int
}

func (s *TokenScanner) Push(token *Token) {
	s.tokens = append(s.tokens, token)
}

func (s *TokenScanner) Last() *Token {
	length := len(s.tokens)
	if length == 0 {
		return nil
	}
	return s.tokens[length-1]
}

func maybeUnquote(str string) string {
	if len(str) > 0 && str[0] == '"' {
		v, err := strconv.Unquote(str)
		if err == nil {
			return v
		}
	}
	return str
}

func (s *TokenScanner) Scan() (Node, error) {
	return nodeFromTokenList(s.tokens)
}

func nodeFromTokenList(tokens []*Token) (Node, error) {
	node := Node(nil)
	label := Node(nil)
	i := 0

	for {
		if i >= len(tokens)-1 {
			break
		}

		t := tokens[i]

		if t.IsKeyword {
			switch t.TokenText {
			case "(":
				nextIdx := i + 1
				depth := 1
				for j, tok := range tokens[nextIdx:] {
					i++
					if tok.TokenText == "(" {
						depth++
					} else if tok.TokenText == ")" {
						depth--
						if depth == 0 {
							nextNode, _ := nodeFromTokenList(tokens[nextIdx : nextIdx+j])
							if node == nil {
								node = nextNode
							} else if expr, ok := node.(*Cond); ok {
								expr.Right = nextNode
							}
							break
						}
					}
				}
			case "!=", "=":
				if i > 0 {
					label = NewLabelWithOperator(tokens[i-1].TokenText, maybeUnquote(tokens[i+1].TokenText), t.TokenText)
					if node == nil {
						node = label
					} else if expr, ok := node.(*Cond); ok {
						expr.Right = label
					}
					i++
				}
			case "AND", "OR":
				op := CondOperatorAND
				if t.TokenText == "OR" {
					op = CondOperatorOR
				}

				node = NewCond(op, node, nil)
			}
		}
		i++
	}

	if expr, ok := node.(*Cond); ok && expr.Right == nil {
		return nil, errors.New("lql syntax error")
	}

	return node, nil
}
