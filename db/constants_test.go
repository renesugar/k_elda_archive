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
	tests := []struct {
		input       string
		expProvider Provider
	}{
		{string(Amazon), Amazon},
		{string(Google), Google},
		{string(DigitalOcean), DigitalOcean},
		{string(Vagrant), Vagrant},
	}

	for _, test := range tests {
		actualProvider, actualErr := ParseProvider(test.input)
		assert.NoError(t, actualErr)
		assert.Equal(t, test.expProvider, actualProvider)
	}
}
