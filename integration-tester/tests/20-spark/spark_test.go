package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	testerUtil "github.com/kelda/kelda/integration-tester/util"
)

func TestCalculatesPI(t *testing.T) {
	clnt, err := testerUtil.GetDefaultDaemonClient()
	if err != nil {
		t.Fatalf("couldn't get api client: %s", err.Error())
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		t.Fatalf("couldn't query containers: %s", err.Error())
	}

	containersPretty, _ := exec.Command("kelda", "ps").Output()
	fmt.Println("`kelda ps` output:")
	fmt.Println(string(containersPretty))

	var id string
	for _, dbc := range containers {
		if strings.HasPrefix(dbc.Hostname, "spark-ms") {
			id = dbc.BlueprintID
			break
		}
	}
	if id == "" {
		t.Fatal("unable to find BlueprintID of Spark master")
	}

	cmd := exec.Command("kelda", "ssh", id, "run-example", "SparkPi")
	output, err := cmd.Output()
	fmt.Println(string(output))
	if err != nil {
		t.Fatalf("Failed to run spark job: %s", err.Error())
	}

	if !strings.Contains(string(output), "Pi is roughly") {
		t.Fatalf("Failed to calculate Pi")
	}
}
