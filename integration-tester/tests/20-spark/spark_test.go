package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	testerUtil "github.com/kelda/kelda/integration-tester/util"
)

func TestRunShuffleJob(t *testing.T) {
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

	// Run a job that shuffles 100,000 key-value pairs using 100 mappers and 10
	// reducers.
	numMappers := 100
	keyValuePairsPerMapper := 1000
	valuesPerKey := 4
	numReducers := 10
	cmd := exec.Command("kelda", "ssh", id, "run-example",
		fmt.Sprintf("GroupByTest %d %d %d %d",
			numMappers, keyValuePairsPerMapper, valuesPerKey, numReducers))
	stderrBytes := bytes.NewBuffer(nil)
	cmd.Stderr = stderrBytes

	stdoutBytes, err := cmd.Output()
	t.Log(stderrBytes.String())
	fmt.Println(string(stdoutBytes))
	if err != nil {
		t.Fatalf("Failed to run spark job: %s", err.Error())
	}

	outputAsStr := strings.Trim(string(stdoutBytes), " \n\t")
	outputAsNum, err := strconv.Atoi(outputAsStr)
	if err != nil {
		t.Fatalf("Output of Spark job was expected to be a number (was %s)",
			outputAsStr)
	}

	// The expected output is the number of unique keys.  Since the keys are randomly
	// generated, there will be some overlap in the keys, so there won't quite be
	// numMappers*numKeyValuePairsPerMapper (i.e., 100,000) unique keys.
	maxOutput := numMappers * keyValuePairsPerMapper
	minOutput := int(float64(maxOutput) * 0.99)
	if outputAsNum < minOutput || outputAsNum > maxOutput {
		t.Fatalf("Spark shuffle job result (%d) was not in the expected range",
			outputAsNum)
	}
}
