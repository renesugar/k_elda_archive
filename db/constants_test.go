package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProvider(t *testing.T) {
	t.Parallel()

	_, err := ParseProvider("not_a_provider")
	assert.Error(t, err)
	expErr := errors.New("provider not_a_provider not supported (supported " +
		"providers: [Amazon Google DigitalOcean Vagrant])")
	assert.Equal(t, expErr, err)

	// Verify that the correct provider is returned for all supported providers.
	for _, provider := range AllProviders {
		actualProvider, err := ParseProvider(string(provider))
		assert.NoError(t, err)
		assert.Equal(t, provider, actualProvider)
	}
}
