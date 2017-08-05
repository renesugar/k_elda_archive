package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlacement(t *testing.T) {
	t.Parallel()

	conn := New()

	var id int
	conn.Txn(PlacementTable).Run(func(view Database) error {
		placement := view.InsertPlacement()
		id = placement.ID
		placement.Size = "foo"
		view.Commit(placement)
		return nil
	})

	placements := PlacementSlice(conn.SelectFromPlacement(
		func(i Placement) bool { return true }))
	assert.Equal(t, 1, placements.Len())

	placement := placements[0]
	assert.Equal(t, "foo", placement.Size)
	assert.Equal(t, id, placement.getID())

	assert.Equal(t, "Placement-1{Exclusive=false, Size=foo}", placement.String())

	assert.Equal(t, placement, placements.Get(0))

	assert.True(t, placement.less(Placement{ID: id + 1}))
}
