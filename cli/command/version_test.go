package command

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kelda/kelda/api/client"
	"github.com/kelda/kelda/api/client/mocks"
	"github.com/kelda/kelda/connection"
	"github.com/kelda/kelda/connection/tls"
)

func TestVersionFlags(t *testing.T) {
	t.Parallel()

	expHost := "mockHost"

	cmd := NewVersionCommand()
	err := parseHelper(cmd, []string{"-H", expHost})

	assert.NoError(t, err)
	assert.Equal(t, expHost, cmd.host)
}

func TestGetDaemonVersion(t *testing.T) {
	t.Parallel()

	mockLocalClient := &mocks.Client{}
	mockLocalClient.On("Version").Once().Return("mockVersion", nil)
	mockLocalClient.On("Close").Return(nil)
	mockGetter := func(_ string, _ connection.Credentials) (client.Client, error) {
		return mockLocalClient, nil
	}
	mockCredsGetter := func(path string) (tls.TLS, error) {
		return tls.TLS{}, nil
	}

	vCmd := NewVersionCommand()

	version, err := vCmd.fetchDaemonVersion(mockCredsGetter, mockGetter)
	assert.Equal(t, "mockVersion", version)
	assert.Nil(t, err)

	mockLocalClient.On("Version").Return("", assert.AnError)
	_, err = vCmd.fetchDaemonVersion(mockCredsGetter, mockGetter)
	assert.NotNil(t, err)
}
