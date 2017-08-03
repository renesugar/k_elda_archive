package vagrant

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quilt/quilt/cloud/machine"
)

func TestSetACLs(t *testing.T) {
	prvdr := Provider{}
	assert.Nil(t, prvdr.SetACLs(nil))
}

func TestPreemptibleError(t *testing.T) {
	err := Provider{}.Boot([]machine.Machine{{Preemptible: true}})
	assert.EqualError(t, err, "vagrant does not support preemptible instances")
}
