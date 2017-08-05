package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinion(t *testing.T) {
	t.Parallel()

	conn := New()

	var id int
	conn.Txn(MinionTable).Run(func(view Database) error {
		minion := view.InsertMinion()
		id = minion.ID
		minion.Blueprint = "foo"
		minion.Self = true
		view.Commit(minion)
		return nil
	})

	minions := MinionSlice(conn.SelectFromMinion(func(i Minion) bool { return true }))
	assert.Equal(t, 1, minions.Len())

	minion := minions[0]
	assert.Equal(t, "foo", minion.Blueprint)
	assert.Equal(t, id, minion.getID())

	assert.Equal(t, "Minion-1{Self=true, HostSubnets=[]}", minion.String())

	assert.Equal(t, minion, minions.Get(0))

	assert.True(t, minion.less(Minion{ID: id + 1}))

	assert.Equal(t, "foo", conn.MinionSelf().Blueprint)
}
