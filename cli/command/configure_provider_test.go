package command

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigProviderFlags(t *testing.T) {
	t.Parallel()

	configProvider := &ConfigProvider{}

	flags := &flag.FlagSet{}
	assert.Nil(t, flags.Usage)

	configProvider.InstallFlags(flags)

	assert.NotNil(t, flags.Usage)
}

func TestConfigProviderParse(t *testing.T) {
	t.Parallel()

	configProvider := &ConfigProvider{}
	assert.NoError(t, configProvider.Parse(nil))

	// Shouldn't error even if passed a redundant argument.
	assert.NoError(t, configProvider.Parse([]string{"foo"}))
}
