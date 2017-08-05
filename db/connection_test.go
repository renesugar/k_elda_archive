package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnection(t *testing.T) {
	t.Parallel()

	conn := New()

	var id int
	conn.Txn(ConnectionTable).Run(func(view Database) error {
		connection := view.InsertConnection()
		id = connection.ID
		connection.From = "foo"
		view.Commit(connection)
		return nil
	})

	connections := ConnectionSlice(conn.SelectFromConnection(
		func(i Connection) bool { return true }))
	assert.Equal(t, 1, connections.Len())

	connection := connections[0]
	assert.Equal(t, "foo", connection.From)
	assert.Equal(t, id, connection.getID())

	connection.MaxPort = 3
	assert.Equal(t, "Connection-1{foo->:0-3}", connection.String())
	connection.MaxPort = 0

	assert.Equal(t, connection, connections.Get(0))

	assert.True(t, connection.less(Connection{From: "z"}))
	assert.True(t, connection.less(Connection{From: "foo", To: "a"}))
	assert.True(t, connection.less(Connection{From: "foo", MaxPort: 1}))
	assert.True(t, connection.less(Connection{From: "foo", MinPort: 100}))
	assert.True(t, connection.less(Connection{From: "foo", ID: id + 1}))
}
