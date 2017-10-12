package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test writes a secret into Vault so that the next test can run a
// blueprint that references a secret. This is necessary because the
// integration-tester waits until all containers have started before running
// any test files. Thus, we can't create the secret in a single test suite.
func TestSecretSetup(t *testing.T) {
	err := runCmd("kelda", "secret", "myEnvSecret", "env secret")
	assert.NoError(t, err)

	err = runCmd("kelda", "secret", "myFileSecret", "file secret")
	assert.NoError(t, err)
}

func runCmd(cmdName string, cmdArgs ...string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
