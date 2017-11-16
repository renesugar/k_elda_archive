package main

import (
	"sync"
	"testing"

	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/integration-tester/util"
)

func TestNetwork(t *testing.T) {
	clnt, err := util.GetDefaultDaemonClient()
	if err != nil {
		t.Fatalf("couldn't get api client: %s", err.Error())
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		t.Fatalf("couldn't query containers: %s", err.Error())
	}

	machines, err := clnt.QueryMachines()
	if err != nil {
		t.Fatalf("failed to query machines: %s", err)
	}

	sshUtil := util.NewSSHUtil(machines)
	t.Run("Ping", func(t *testing.T) {
		connTester, err := newConnectionTester(clnt, sshUtil)
		if err != nil {
			t.Fatalf("couldn't initialize connection tester: %s",
				err.Error())
		}

		testContainers(t, connTester, containers)
	})

	t.Run("DNS", func(t *testing.T) {
		dnsTester, err := newDNSTester(clnt, sshUtil)
		if err != nil {
			t.Fatalf("couldn't initialize dns tester: %s", err.Error())
		}

		testContainers(t, dnsTester, containers)
	})
}

type testerIntf interface {
	test(t *testing.T, c db.Container)
}

// Gather test results for each container. For each minion machine, run one test
// at a time.
func testContainers(t *testing.T, tester testerIntf, containers []db.Container) {
	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, c := range containers {
		go func(c db.Container) {
			tester.test(t, c)
			wg.Done()
		}(c)
	}
	wg.Wait()
}
