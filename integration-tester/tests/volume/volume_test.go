package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/kelda/kelda/integration-tester/util"

	"github.com/stretchr/testify/assert"
)

const hasMountPrefix = "has-mount"
const noMountPrefix = "no-mount"
const mountedPath = "/docker.sock"

func TestVolume(t *testing.T) {
	client, err := util.GetDefaultDaemonClient()
	assert.NoError(t, err)
	defer client.Close()

	containers, err := client.QueryContainers()
	assert.NoError(t, err)

	for _, dbc := range containers {
		switch {
		case strings.HasPrefix(dbc.Hostname, hasMountPrefix):
			assert.NoError(t, tryLs(dbc.Hostname, mountedPath))
		case strings.HasPrefix(dbc.Hostname, noMountPrefix):
			assert.NotNil(t, tryLs(dbc.Hostname, mountedPath))
		default:
			fmt.Printf("unexpected container: %s\n", dbc.Hostname)
		}
	}
}

func tryLs(container, path string) error {
	fmt.Printf("Listing %s on %s..\n", path, container)
	cmd := exec.Command("kelda", "ssh", container, "ls", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to ls %s (%s): %s", path, err, output)
	}
	return nil
}
