package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kelda/kelda/cli/ssh"
	"github.com/kelda/kelda/db"
)

// The maximum number of SSH Sessions spread across all minions.
const maxSSHSessions = 512

// SSHUtil makes it easy to parallelize executing commands on Kelda containers.
//
// It does several things to make executing commands faster:
// 1. It reuses SSH connections.
//
// 2. It caches state. Shelling out to `kelda ssh` requires requerying all
// machines and containers in order to figure out the container's public IP
// address.
//
// 3. It rate limits the number of parallel connections and session creations
// to avoid overloading the host.
type SSHUtil struct {
	clients map[string]chan ssh.Client
}

// NewSSHUtil creates a new SSHUtil instance configured to connect to containers
// on the given machines. Any calls to `SSHUtil.SSH` for a container scheduled
// on a machine not given here will fail.
func NewSSHUtil(machines []db.Machine) SSHUtil {
	sshUtil := SSHUtil{map[string]chan ssh.Client{}}

	// Map writes aren't thread safe, so we create the channels before go
	// routines have a chance to launch.
	for _, m := range machines {
		sshUtil.clients[m.PrivateIP] = make(chan ssh.Client, maxSSHSessions)
	}

	// We have to limit parallelization setting up SSH sessions.  Doing so too
	// quickly in parallel breaks system-logind on the remote machine:
	// https://github.com/systemd/systemd/issues/2925.  Furthermore, the concurrency
	// limit cannot exceed the sshd MaxStartups setting, or else the SSH connections
	// may be randomly rejected.
	//
	// Also note, we intentionally don't wait for this go routine to finish.  As new
	// SSH connections are created, the tests can gradually take advantage of them.
	go func() {
		for i := 0; i < maxSSHSessions; i++ {
			m := machines[i%len(machines)]
			client, err := ssh.New(m.PublicIP, "")
			if err != nil {
				fmt.Printf("failed to ssh to %s: %s", m.PublicIP, err)
				continue
			}
			sshUtil.clients[m.PrivateIP] <- client
		}
	}()
	return sshUtil
}

// SSH executes `cmd` on the given container, and returns the stdout and stderr
// output of the command in a single string.
func (sshUtil SSHUtil) SSH(dbc db.Container, cmd ...string) (string, error) {
	if dbc.Minion == "" || dbc.DockerID == "" {
		return "", errors.New("container not yet booted")
	}

	sshChan, ok := sshUtil.clients[dbc.Minion]
	if !ok {
		return "", fmt.Errorf("unknown machine: %s", dbc.Minion)
	}

	ssh := <-sshChan
	defer func() { sshChan <- ssh }()

	fmt.Println(dbc.Hostname, strings.Join(cmd, " "))
	ret, err := ssh.CombinedOutput(fmt.Sprintf("docker exec %s %s", dbc.DockerID,
		strings.Join(cmd, " ")))
	return string(ret), err
}
