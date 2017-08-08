package command

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitFlags(t *testing.T) {
	t.Parallel()

	init := &Init{}

	flags := &flag.FlagSet{}
	assert.Nil(t, flags.Usage)

	init.InstallFlags(flags)

	assert.NotNil(t, flags.Usage)
}

func TestInitParse(t *testing.T) {
	t.Parallel()

	init := &Init{}
	assert.NoError(t, init.Parse(nil), "")

	// Shouldn't error even if passed a redundant argument.
	assert.NoError(t, init.Parse([]string{"foo"}))
}
