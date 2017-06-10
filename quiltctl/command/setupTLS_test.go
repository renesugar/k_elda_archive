package command

import (
	"errors"
	"testing"

	"github.com/quilt/quilt/connection/credentials/tls"
	"github.com/quilt/quilt/quiltctl/command/credentials"
	"github.com/quilt/quilt/util"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// Test that the generated files can be parsed into a connection.Credential.
func TestSetupTLS(t *testing.T) {
	util.AppFs = afero.NewMemMapFs()

	cmd := SetupTLS{outDir: "tls"}
	exitCode := cmd.Run()
	assert.Zero(t, exitCode)

	creds, err := credentials.Read("tls")
	assert.NoError(t, err)

	_, ok := creds.(tls.TLS)
	assert.True(t, ok)
}

func checkSetupTLSParsing(t *testing.T, args []string, exp string, expErr error) {
	cmd := &SetupTLS{}
	err := parseHelper(cmd, args)

	assert.Equal(t, expErr, err)
	assert.Equal(t, exp, cmd.outDir)
}

func TestSetupTLSFlags(t *testing.T) {
	t.Parallel()

	exp := "foo"
	checkSetupTLSParsing(t, []string{"-out", exp}, exp, nil)
	checkSetupTLSParsing(t, []string{exp}, exp, nil)
	checkSetupTLSParsing(t, []string{}, "",
		errors.New("no output directory specified"))
}
