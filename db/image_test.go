package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImage(t *testing.T) {
	t.Parallel()

	conn := New()

	var id int
	conn.Txn(ImageTable).Run(func(view Database) error {
		image := view.InsertImage()
		id = image.ID
		image.Name = "foo"
		view.Commit(image)
		return nil
	})

	images := ImageSlice(conn.SelectFromImage(func(i Image) bool { return true }))
	assert.Equal(t, 1, images.Len())

	image := images[0]
	assert.Equal(t, "foo", image.Name)
	assert.Equal(t, id, image.getID())
	assert.Equal(t, ImageTable, image.tt())

	assert.Equal(t, "Image-1{Name=foo}", image.String())

	assert.Equal(t, image, images.Get(0))

	assert.True(t, image.less(Image{ID: id + 1}))
}
