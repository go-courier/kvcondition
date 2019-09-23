package kvcondition

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkParseKVCondition(b *testing.B) {
	rule := []byte(`tag = ONLINE & tag = "some label" & ( ip != 1.1.1.1 | ip ^= "8.8" | tag = "test&" ) | ip = 4.4.4.4`)

	for i := 0; i < b.N; i++ {
		ParseKVCondition(rule)
	}
}

func TestParseKVCondition(t *testing.T) {
	tt := require.New(t)

	rule := `tag = ONLINE & tag = "some label" & ( ip != 1.1.1.1 | ip ^= "8.8" | tag = "test&" ) | ip = 4.4.4.4`

	node, err := ParseKVCondition([]byte(rule))
	tt.NoError(err)

	tt.Equal(
		`( ( ( tag = "ONLINE" & tag = "some label" ) & ( ( ip != "1.1.1.1" | ip ^= "8.8" ) | tag = "test&" ) ) | ip = "4.4.4.4" )`,
		node.String(),
	)

	kvc, err := ParseKVCondition([]byte(node.String()))
	tt.NoError(err)
	tt.Equal(node.String(), kvc.String())

	rules := make([]*Rule, 0)

	kvc.Range(func(label *Rule) {
		rules = append(rules, label)
	})

	tt.Equal([]*Rule{
		OperatorEqual.Of("tag", "ONLINE"),
		OperatorEqual.Of("tag", "some label"),
		OperatorNotEqual.Of("ip", "1.1.1.1"),
		OperatorStartsWith.Of("ip", "8.8"),
		OperatorEqual.Of("tag", "test&"),
		OperatorEqual.Of("ip", "4.4.4.4"),
	}, rules)
}

func TestParseKVConditionFailed(t *testing.T) {
	tt := require.New(t)

	rule := "tag = ONLINE & tag = \"some label\" & ( ip = 1.1.1.1 | ip = \"8.8.8.8\" | tag = test & ip = 4.4.4.4"

	_, err := ParseKVCondition([]byte(rule))
	tt.Error(err)
}
