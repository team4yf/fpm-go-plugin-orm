package pg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
	req := &queryReq{
		Table:     "fake",
		Condition: "name = 'C'",
		Fields:    "name,val",
		Skip:      0,
		Limit:     0,
		Data:      "",
		ID:        0,
		Sort:      "id-",
	}
	q := parseQuery(req)
	assert.Equal(t, q.Table, "fake", "shoule be fake")
	assert.Equal(t, q.Condition, "name = 'C'", "")
	assert.Equal(t, q.Pager.Skip, 0, "")
	assert.Equal(t, q.Fields[0], "name", "")
	assert.Equal(t, q.Fields[1], "val", "")
	assert.Equal(t, q.Pager.Limit, -1, "")
	assert.Equal(t, q.Sorter[0].Sortby, "id", "")
	assert.Equal(t, q.Sorter[0].Asc, "desc", "")
}
