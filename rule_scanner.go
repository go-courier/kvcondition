package kvcondition

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"text/scanner"
)

var keywords = map[rune]bool{
	'!': true,
	'=': true,
	'(': true,
	')': true,
	'*': true,
	'^': true,
	'$': true,
	'&': true,
	'|': true,
}

func newRuleScanner(b []byte) *ruleScanner {
	s := &scanner.Scanner{}
	s.Init(bytes.NewReader(b))

	return &ruleScanner{
		data:    b,
		Scanner: s,
	}
}

type ruleScanner struct {
	data []byte
	*scanner.Scanner
}

func (s *ruleScanner) ScanNode() (Node, error) {
	node := Node(nil)
	rule := (*Rule)(nil)

	token := bytes.NewBuffer(nil)
	op := bytes.NewBuffer(nil)

	completeNode := func(nextNode Node) {
		switch n := node.(type) {
		case *Condition:
			n.Right = nextNode
		default:
			node = nextNode
		}
	}

	completeRule := func() {
		if rule != nil {
			rule.Value = []byte(maybeUnquote(strings.TrimSpace(token.String())))
			rule = nil
			token = bytes.NewBuffer(nil)
		}
	}

	for tok := s.Next(); ; tok = s.Next() {
		if tok == scanner.EOF {
			break
		}

		if tok == ')' {
			break
		}

		// quote skip
		if tok == '"' {
			for tok := s.Next(); tok != '"'; tok = s.Next() {
				if tok == '\\' {
					tok = s.Next()
				}
				token.WriteRune(tok)
			}
			tok = s.Next()
		}

		if keywords[tok] {
			completeRule()

			switch tok {
			case '(':
				nextNode, err := s.ScanNode()
				if err != nil {
					return nil, err
				}
				completeNode(nextNode)
			case '&':
				node = And(node, nil)
			case '|':
				node = Or(node, nil)
			case '=':
				op.WriteRune(tok)

				rule = &Rule{Key: bytes.TrimSpace(token.Bytes())}

				token = bytes.NewBuffer(nil)

				switch op.String() {
				case "=":
					rule.Operator = OperatorEqual
				case "!=":
					rule.Operator = OperatorNotEqual
				case "*=":
					rule.Operator = OperatorContains
				case "^=":
					rule.Operator = OperatorStartsWith
				case "$=":
					rule.Operator = OperatorEndsWith
				}

				op = bytes.NewBuffer(nil)
				completeNode(rule)
			default:
				op.WriteRune(tok)
			}

			continue
		}

		// collect key or value
		token.WriteRune(tok)
	}

	completeRule()

	if expr, ok := node.(*Condition); ok && expr.Right == nil {
		return nil, errors.New("kv condition syntax error")
	}

	return node, nil
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
