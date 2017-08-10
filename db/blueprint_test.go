package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBlueprintNamespace(t *testing.T) {
	conn := New()

	ns, err := conn.GetBlueprintNamespace()
	assert.NotNil(t, err)
	assert.Exactly(t, ns, "")

	conn.Txn(AllTables...).Run(func(view Database) error {
		bp := view.InsertBlueprint()
		bp.Namespace = "test"
		view.Commit(bp)
		return nil
	})

	ns, err = conn.GetBlueprintNamespace()
	assert.NoError(t, err)
	assert.Exactly(t, ns, "test")
}
