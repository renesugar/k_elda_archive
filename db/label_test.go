package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel(t *testing.T) {
	t.Parallel()

	conn := New()

	var id int
	conn.Txn(LabelTable).Run(func(view Database) error {
		label := view.InsertLabel()
		id = label.ID
		label.Label = "foo"
		view.Commit(label)
		return nil
	})

	labels := LabelSlice(conn.SelectFromLabel(
		func(i Label) bool { return true }))
	assert.Equal(t, 1, labels.Len())

	label := labels[0]
	assert.Equal(t, "foo", label.Label)
	assert.Equal(t, id, label.getID())

	assert.Equal(t, "Label-1{Label=foo, ContainerIPs=[]}", label.String())

	assert.Equal(t, label, labels.Get(0))

	assert.True(t, label.less(Label{Label: "z"}))
	assert.True(t, label.less(Label{Label: "foo", ID: id + 1}))
}
