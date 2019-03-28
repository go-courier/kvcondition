package labelql

import (
	"encoding"
	"strconv"
	"strings"
	"text/scanner"
)

func ParseLabelQL(lql string) (*LabelQL, error) {
	s := scanner.Scanner{}
	s.Init(strings.NewReader(lql))

	tokenScanner := TokenScanner{}

	for {
		tok := s.Scan()
		if tok == scanner.EOF {
			break
		}

		nextToken := NewToken(s.Pos(), s.TokenText())

		lastToken := tokenScanner.Last()
		if lastToken != nil {
			if !lastToken.IsKeyword && !nextToken.IsKeyword {
				lastToken.TokenText += nextToken.TokenText
				continue
			}

			if lastToken.TokenText == "!" && nextToken.TokenText == "=" {
				lastToken.TokenText += nextToken.TokenText
				continue
			}
		}
		tokenScanner.Push(nextToken)
	}

	node, err := tokenScanner.Scan()
	if err != nil {
		return nil, err
	}

	labelQL := &LabelQL{
		Node: node,
	}

	return labelQL, nil
}

// swagger:strfmt lql
type LabelQL struct {
	Node
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*LabelQL)(nil)

func (v LabelQL) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *LabelQL) UnmarshalText(data []byte) error {
	labelQL, err := ParseLabelQL(string(data))
	if err != nil {
		return err
	}
	*v = *labelQL
	return nil
}

type labelQLVisit func(visit labelQLVisit, node Node)

func (v *LabelQL) RangeLabel(cb func(label *Label)) {
	visit := func(next labelQLVisit, node Node) {
		if c, ok := node.(*Cond); ok {
			next(next, c.Left)
			next(next, c.Right)
			return
		}
		if label, ok := node.(*Label); ok {
			cb(label)
		}
	}
	visit(visit, v.Node)
}

func (v *LabelQL) Match(labels []*Label) bool {
	return isNodeMatchLabels(v.Node, labels)
}

type Node interface {
	String() string
}

func NewLabel(key string, value string) *Label {
	return &Label{
		Operator: "=",
		Key:      key,
		Value:    value,
	}
}

func NewLabelWithOperator(key string, value string, op string) *Label {
	return &Label{
		Operator: op,
		Key:      key,
		Value:    value,
	}
}

type Label struct {
	Key      string
	Value    string
	Operator string
}

func (l *Label) Match(labels []*Label) bool {
	otherwise := false
	if l.Operator != "=" {
		otherwise = true
	}

	for _, label := range labels {
		if l.Equal(label) {
			return !otherwise
		}
	}

	return otherwise
}

func (l *Label) Equal(targetLabel *Label) bool {
	if l == nil || targetLabel == nil {
		return false
	}
	return l.Key == targetLabel.Key && strings.ToUpper(l.Value) == strings.ToUpper(targetLabel.Value)
}

func (l *Label) String() string {
	return l.Key + " " + l.Operator + " " + strconv.Quote(l.Value)
}

func NewCond(op CondOperator, left Node, right Node) *Cond {
	return &Cond{
		Operator: op,
		Left:     left,
		Right:    right,
	}
}

type CondOperator int

const (
	CondOperatorAND CondOperator = iota + 1
	CondOperatorOR
)

type Cond struct {
	Operator CondOperator
	Left     Node
	Right    Node
}

func isNodeMatchLabels(node Node, labels []*Label) bool {
	if node == nil || len(labels) == 0 {
		return false
	}

	switch node.(type) {
	case *Label:
		label := node.(*Label)
		return label.Match(labels)
	case *Cond:
		cond := node.(*Cond)
		return cond.Match(labels)
	}

	return false
}

func (c *Cond) Match(labels []*Label) bool {
	leftMatch := isNodeMatchLabels(c.Left, labels)
	rightMatch := isNodeMatchLabels(c.Right, labels)
	if c.Operator == CondOperatorAND {
		return leftMatch && rightMatch
	}
	if c.Operator == CondOperatorOR {
		return leftMatch || rightMatch
	}
	return false
}

func (c *Cond) String() string {
	op := ""
	switch c.Operator {
	case CondOperatorAND:
		op = " AND "
	case CondOperatorOR:
		op = " OR "
	}

	return "( " + c.Left.String() + op + c.Right.String() + " )"
}
