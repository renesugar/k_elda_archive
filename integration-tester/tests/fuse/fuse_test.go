package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	canMountName    = "can-mount"
	cannotMountName = "cannot-mount"
	sshServerName   = "ssh-server"
)

func TestSSHFs(t *testing.T) {
	// Assert that both containers can SSH.
	sshCmd := []string{"ssh", "-o", "StrictHostKeyChecking=no", sshServerName,
		"true"}
	assert.NoError(t, tryExec(canMountName, sshCmd))
	assert.NoError(t, tryExec(cannotMountName, sshCmd))

	// But only the privileged container can create a sshfs mount.
	mountCmd := []string{"sshfs", fmt.Sprintf("%s:/", sshServerName), "/mnt"}
	assert.NoError(t, tryExec(canMountName, mountCmd))
	assert.NotNil(t, tryExec(cannotMountName, mountCmd))

	// Test that the mount is really from the ssh-server container.
	mountedHostname, err := catFile(canMountName, "/mnt/etc/hostname")
	assert.NoError(t, err)
	assert.Equal(t, sshServerName, strings.TrimRight(mountedHostname, "\n"))
}

func tryExec(container string, cmd []string) error {
	execCmd := exec.Command("kelda", append([]string{"ssh", container}, cmd...)...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to exec %v (%s): %s", cmd, err, output)
	}
	return nil
}

func catFile(container, path string) (string, error) {
	cmd := exec.Command("kelda", "ssh", container, "cat", path)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
