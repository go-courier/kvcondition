package kvcondition

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkParseKVCondition(b *testing.B) {
	rule := []byte(`env && tag = ONLINE & tag = "some label" & ip & ( ip != 1.1.1.1 | ip ^= "8.8" | tag *= "test&" ) | ip $= 4.4`)

	for i := 0; i < b.N; i++ {
		ParseKVCondition(rule)
	}
}

func TestKVCondition(t *testing.T) {
	ql, _ := ParseKVCondition([]byte(`ip != "1.1.1.1"`))

	type Data struct {
		QL KVCondition `json:"ql"`
	}

	data, err := json.Marshal(&Data{
		QL: *ql,
	})
	require.NoError(t, err)

	d := Data{}
	er := json.Unmarshal(data, &d)
	require.NoError(t, er)

	require.True(t, d.QL.String() == ql.String())
}

func TestParseKVCondition(t *testing.T) {
	tt := require.New(t)

	rule := []byte(`env & tag = ONLINE & tag = "some label" & ip & ( ip != 1.1.1.1 | ip ^= "8.8" | tag *= "test\&" ) | ip $= 4.4`)

	kvCondition := &KVCondition{}
	err := kvCondition.UnmarshalText(rule)
	tt.NoError(err)

	tt.Equal(
		`( ( ( ( ( env & tag = "ONLINE" ) & tag = "some label" ) & ip ) & ( ( ip != "1.1.1.1" | ip ^= "8.8" ) | tag *= "test&" ) ) | ip $= "4.4" )`,
		kvCondition.Node.String(),
	)

	kvc, err := ParseKVCondition([]byte(kvCondition.Node.String()))
	tt.NoError(err)
	tt.Equal(kvCondition.String(), kvc.String())

	rules := make([]*Rule, 0)

	kvc.Range(func(label *Rule) {
		rules = append(rules, label)
	})

	tt.Equal([]*Rule{
		OperatorExists.Of("env", ""),
		OperatorEqual.Of("tag", "ONLINE"),
		OperatorEqual.Of("tag", "some label"),
		OperatorExists.Of("ip", ""),
		OperatorNotEqual.Of("ip", "1.1.1.1"),
		OperatorStartsWith.Of("ip", "8.8"),
		OperatorContains.Of("tag", "test&"),
		OperatorEndsWith.Of("ip", "4.4"),
	}, rules)
}

func TestParseKVConditionFailed(t *testing.T) {
	tt := require.New(t)

	rule := "tag = ONLINE & tag = \"some label\" & ( ip = 1.1.1.1 | ip = \"8.8.8.8\" | tag = test & ip = 4.4.4.4"

	_, err := ParseKVCondition([]byte(rule))
	tt.Error(err)
}
