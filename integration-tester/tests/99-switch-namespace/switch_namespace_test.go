package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/kelda/kelda/api/client"
	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	testerUtil "github.com/kelda/kelda/integration-tester/util"
	"github.com/kelda/kelda/util"

	"github.com/stretchr/testify/assert"
)

// TestSwitchNamespace tests that if a user switches namespaces then runs the
// same blueprint, containers are not restarted.
func TestSwitchNamespace(t *testing.T) {
	clnt, err := testerUtil.GetDefaultDaemonClient()
	assert.NoError(t, err)

	initialBlueprint, err := getCurrentBlueprint(clnt)
	assert.NoError(t, err)

	initialContainers, err := clnt.QueryContainers()
	assert.NoError(t, err)

	fmt.Println("Simulating switching namespaces by running `kelda stop`")
	cmd := exec.Command("kelda", "stop", "random-namespace")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assert.NoError(t, cmd.Run())

	// Wait until the daemon has switched namespaces by waiting for it to pick
	// up that there are no machines in the new namespace.
	switchedNamespace := func() bool {
		machines, err := clnt.QueryMachines()
		if err != nil {
			fmt.Println("Failed to list machines:", err)
			return false
		}
		return len(machines) == 0
	}
	err = util.BackoffWaitFor(switchedNamespace, 30*time.Second, 3*time.Minute)
	assert.NoError(t, err)

	fmt.Println("Switching back to the original blueprint")
	assert.NoError(t, clnt.Deploy(initialBlueprint.String()))

	// Wait until the daemon has redeployed the initial blueprint (i.e. wait
	// until the specified containers are running).
	containersUp := func() bool {
		containers, err := clnt.QueryContainers()
		if err != nil {
			fmt.Println("Failed to list containers:", err)
			return false
		}

		// If the DockerID isn't set, either the container has not been started
		// yet, or the query response did not contact the Worker. Either way,
		// the container response is incomplete, so we should wait until the
		// container is booted, or the daemon connects to the worker.
		for _, c := range containers {
			if c.DockerID == "" {
				return false
			}
		}
		return len(containers) == len(initialBlueprint.Containers)
	}
	err = util.BackoffWaitFor(containersUp, 30*time.Second, 3*time.Minute)
	assert.NoError(t, err)

	// Ensure that the running containers are exactly the same as before we
	// switched namespaces.
	currentContainers, err := clnt.QueryContainers()
	assert.NoError(t, err)

	fmt.Printf("Initial containers: %+v\n", initialContainers)
	fmt.Printf("Current containers: %+v\n", currentContainers)
	assert.Len(t, currentContainers, len(initialContainers))
	assert.Subset(t, scrubIDs(currentContainers), scrubIDs(initialContainers))
}

func scrubIDs(dbcs []db.Container) (scrubbed []db.Container) {
	for _, c := range dbcs {
		c.ID = 0
		scrubbed = append(scrubbed, c)
	}
	return scrubbed
}

func getCurrentBlueprint(c client.Client) (blueprint.Blueprint, error) {
	bps, err := c.QueryBlueprints()
	if err != nil {
		return blueprint.Blueprint{}, err
	}

	if len(bps) != 1 {
		return blueprint.Blueprint{}, errors.New(
			"unexpected number of blueprints")
	}
	return bps[0].Blueprint, nil
}
