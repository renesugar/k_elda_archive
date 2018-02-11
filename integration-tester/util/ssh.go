package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kelda/kelda/cli/ssh"
	"github.com/kelda/kelda/db"
)

// The maximum number of SSH sessions to concurrently allow to a single machine.
const maxSSHSessions = 100

// SSHUtil makes it easy to parallelize executing commands on Kelda containers.
//
// It does several things to make executing commands faster:
// 1. It reuses SSH connections.
//
// 2. It caches state. Shelling out to `kelda ssh` requires requerying all
// machines and containers in order to figure out the container's public IP
// address.
//
// 3. It rate limits the number of parallel sessions to avoid overloading the
// remote host.
type SSHUtil map[string]sshClient

type sshClient struct {
	ssh.Client
	semaphore chan struct{}
}

// NewSSHUtil creates a new SSHUtil instance configured to connect to containers
// on the given machines. Any calls to `SSHUtil.SSH` for a container scheduled
// on a machine not given here will fail.
func NewSSHUtil(machines []db.Machine) (SSHUtil, error) {
	sshUtil := SSHUtil(map[string]sshClient{})

	for _, m := range machines {
		semaphore := make(chan struct{}, maxSSHSessions)
		for i := 0; i < maxSSHSessions; i++ {
			semaphore <- struct{}{}
		}

		client, err := ssh.New(m.PublicIP, "")
		if err != nil {
			return SSHUtil{}, fmt.Errorf("failed to ssh to %s: %s",
				m.PublicIP, err)
		}
		sshUtil[m.PrivateIP] = sshClient{Client: client, semaphore: semaphore}
	}

	return sshUtil, nil
}

// SSH executes `cmd` on the given container, and returns the stdout and stderr
// output of the command in a single string.
func (sshUtil SSHUtil) SSH(dbc db.Container, cmd ...string) (string, error) {
	if dbc.Minion == "" || dbc.DockerID == "" {
		return "", errors.New("container not yet booted")
	}

	ssh, ok := sshUtil[dbc.Minion]
	if !ok {
		return "", fmt.Errorf("unknown machine: %s", dbc.Minion)
	}

	<-ssh.semaphore
	defer func() { ssh.semaphore <- struct{}{} }()

	fmt.Println(dbc.Hostname, strings.Join(cmd, " "))
	ret, err := ssh.CombinedOutput(fmt.Sprintf("docker exec %s %s", dbc.DockerID,
		strings.Join(cmd, " ")))
	return string(ret), err
}
