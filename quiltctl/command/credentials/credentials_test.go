package credentials

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quilt/quilt/connection/credentials"
	"github.com/quilt/quilt/connection/credentials/tls"
)

func TestGetInsecure(t *testing.T) {
	t.Parallel()

	// Test that no directory implies an insecure connection.
	creds, err := Read("")
	assert.NoError(t, err)
	_, isInsecure := creds.(credentials.Insecure)
	assert.True(t, isInsecure)
}

func TestGetTLS(t *testing.T) {
	// Test that errors are handled.
	tlsFromFile = func(_ string) (tls.TLS, error) {
		return tls.TLS{}, assert.AnError
	}
	_, err := Read("tlsdir")
	assert.NotNil(t, err)

	// Test that we properly propagate the resulting credentials.
	passed := tls.TLS{}
	tlsFromFile = func(_ string) (tls.TLS, error) {
		return passed, nil
	}
	creds, err := Read("tlsdir")
	assert.NoError(t, err)
	assert.Equal(t, passed, creds)
}
