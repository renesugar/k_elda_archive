package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test writes AWS key secrets into Vault so that the next test can run a
// blueprint that references a secret (the test assumes that the environment variables
// AWS_S3_ACCESS_KEY_ID and AWS_S3_SECRET_ACCESS_KEY have been set on the machine where
// this test runs, which is handled by the Jenkins configuration code).
//
// This code needs to be in its own test because the integration-tester
// waits until all containers have started before running any test files. Thus, we
// can't create the secret in a single test suite.
func TestSecretSetup(t *testing.T) {
	awsAccessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_ID")
	err := runCmd("kelda", "secret", "awsAccessKeyID", awsAccessKeyID)
	assert.NoError(t, err)

	awsSecretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY")
	err = runCmd("kelda", "secret", "awsSecretAccessKey", awsSecretAccessKey)
	assert.NoError(t, err)
}

func runCmd(cmdName string, cmdArgs ...string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
