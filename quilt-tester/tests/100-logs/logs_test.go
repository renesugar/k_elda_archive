package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
)

func TestMinionLogs(t *testing.T) {
	if err := printQuiltPs(); err != nil {
		t.Errorf("failed to print quilt ps: %s", err.Error())
	}

	c, err := client.New(api.DefaultSocket)
	if err != nil {
		t.Fatalf("couldn't get quiltctl client: %s", err.Error())
	}
	defer c.Close()

	machines, err := c.QueryMachines()
	if err != nil {
		t.Fatalf("couldn't query the machines: %s", err.Error())
	}

	for _, machine := range machines {
		fmt.Println(machine)
		logsOutput, err := exec.Command("quilt", "ssh", machine.StitchID,
			"sudo", "journalctl", "-o", "cat", "-u", "minion").
			CombinedOutput()
		if err != nil {
			t.Errorf("unable to get minion logs: %s", err.Error())
			continue
		}
		outputStr := string(logsOutput)
		fmt.Println(outputStr)

		// "goroutine 0" is the main goroutine, and is thus always printed in
		// stacktraces.
		// The trailing open bracket is necessary to filter out false positives
		// in the log output. For example, as part of logging DNS requests, the
		// status string NOERROR is printed.
		if strings.Contains(outputStr, "goroutine 0") ||
			strings.Contains(outputStr, "ERROR [") ||
			strings.Contains(outputStr, "WARN [") {
			t.Fatal("minion has error logs")
		}
	}
}

func printQuiltPs() error {
	psout, err := exec.Command("quilt", "ps").CombinedOutput()
	fmt.Println(string(psout))
	return err
}
