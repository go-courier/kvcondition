package labelql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLabelQL(t *testing.T) {
	tt := require.New(t)

	lql := "tag = ONLINE and tag = \"some label\" AND ( ip != 1.1.1.1 or ip = \"8.8.8.8\" or tag = test) OR ip = 4.4.4.4"

	node, err := ParseLabelQL(lql)
	tt.NoError(err)

	tt.Equal(
		`( ( ( tag = "ONLINE" AND tag = "some label" ) AND ( ( ip != "1.1.1.1" OR ip = "8.8.8.8" ) OR tag = "test" ) ) OR ip = "4.4.4.4" )`,
		node.String(),
	)

	labelQL, err := ParseLabelQL(node.String())
	tt.NoError(err)
	tt.Equal(node.String(), labelQL.String())

	labels := make([]*Label, 0)
	labelQL.RangeLabel(func(label *Label) {
		labels = append(labels, label)
	})

	tt.Equal([]*Label{
		NewLabel("tag", "ONLINE"),
		NewLabel("tag", "some label"),
		NewLabelWithOperator("ip", "1.1.1.1", "!="),
		NewLabel("ip", "8.8.8.8"),
		NewLabel("tag", "test"),
		NewLabel("ip", "4.4.4.4"),
	}, labels)

	tt.False(labelQL.Match([]*Label{
		NewLabel("tag", "ONLINE"),
		NewLabel("tag", "some label"),
		NewLabel("ip", "1.1.1.1"),
	}))

	tt.True(labelQL.Match([]*Label{
		NewLabel("tag", "ONLINE"),
		NewLabel("tag", "some label"),
		NewLabel("tag", "test"),
	}))

	tt.True(labelQL.Match([]*Label{
		NewLabel("ip", "4.4.4.4"),
	}))

	tt.False(labelQL.Match([]*Label{
		NewLabel("ip", "8.8.8.8"),
	}))
}

func TestParseLabelQLFailed(t *testing.T) {
	tt := require.New(t)

	lql := "tag = ONLINE and tag = \"some label\" AND ( ip = 1.1.1.1 or ip = \"8.8.8.8\" or tag = test AND ip = 4.4.4.4"

	_, err := ParseLabelQL(lql)
	tt.Error(err)
}
