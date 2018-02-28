package main

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

const privilegedContainerName = "privileged"
const nonPrivilegedContainerName = "not-privileged"

func TestPrivileged(t *testing.T) {
	// Assert that the privileged container can access the host's devices by
	// listing /dev/fuse from within the privileged container.
	assert.NoError(t, tryLsDevFuse(privilegedContainerName))

	// As a sanity check, make sure that listing /dev/fuse on a non-privileged
	// container fails.
	assert.NotNil(t, tryLsDevFuse(nonPrivilegedContainerName))
}

func tryLsDevFuse(container string) error {
	cmd := exec.Command("kelda", "ssh", container, "ls", "/dev/fuse")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to ls /dev/fuse (%s): %s", err, output)
	}
	return nil
}
