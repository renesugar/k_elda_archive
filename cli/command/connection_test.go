package command

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kelda/kelda/api"
	"github.com/kelda/kelda/api/client"
	"github.com/kelda/kelda/api/client/mocks"
	"github.com/kelda/kelda/connection"
)

func TestSetupClient(t *testing.T) {
	t.Parallel()

	// Test that we obtain a client, and properly save it.
	expClient := &mocks.Client{}
	newClient := func(host string, _ connection.Credentials) (client.Client, error) {
		assert.Equal(t, "host", host)
		return expClient, nil
	}
	cmd := connectionHelper{
		connectionFlags: connectionFlags{
			host: "host",
		},
	}
	err := cmd.setupClient(newClient)
	assert.NoError(t, err)
	assert.Equal(t, expClient, cmd.client)

	// Test that errors obtaining a client are properly propagated.
	newClient = func(host string, _ connection.Credentials) (client.Client, error) {
		assert.Equal(t, "host", host)
		return nil, assert.AnError
	}
	cmd = connectionHelper{
		connectionFlags: connectionFlags{
			host: "host",
		},
	}
	err = cmd.setupClient(newClient)
	assert.NotNil(t, err)
}

func TestConnectionFlagsHostEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		keldaHostEnv string
		args         []string
		expHost      string
	}{
		// Test setting the host through an environment variable.
		{
			keldaHostEnv: "envSocket",
			expHost:      "envSocket",
		},

		// Test setting the host through command line arguments.
		{
			args:    []string{"-H", "cmdSocket"},
			expHost: "cmdSocket",
		},

		// Test that command line arguments take precedence.
		{
			keldaHostEnv: "envSocket",
			args:         []string{"-H", "cmdSocket"},
			expHost:      "cmdSocket",
		},

		// Test that when no socket is provided, we use the default.
		{
			expHost: api.DefaultSocket,
		},
	}

	for _, test := range tests {
		os.Setenv("KELDA_HOST", test.keldaHostEnv)
		flags := flag.NewFlagSet("test", flag.ExitOnError)
		cf := connectionFlags{}
		cf.InstallFlags(flags)
		flags.Parse(test.args)
		assert.Equal(t, test.expHost, cf.host)
		os.Setenv("KELDA_HOST", "")
	}
}
