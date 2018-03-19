package command

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseInfraFlags(t *testing.T) {
	t.Parallel()

	baseInfra := &BaseInfra{}

	flags := &flag.FlagSet{}
	assert.Nil(t, flags.Usage)

	baseInfra.InstallFlags(flags)

	assert.NotNil(t, flags.Usage)
}

func TestBaseInfraParse(t *testing.T) {
	t.Parallel()

	baseInfra := &BaseInfra{}
	assert.NoError(t, baseInfra.Parse(nil))

	// Shouldn't error even if passed a redundant argument.
	assert.NoError(t, baseInfra.Parse([]string{"foo"}))
}
