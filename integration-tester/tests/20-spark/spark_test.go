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

	containersPretty, _ := exec.Command("kelda", "ps").Output()
	fmt.Println("`kelda ps` output:")
	fmt.Println(string(containersPretty))

	// Run a job that shuffles 100,000 key-value pairs using 100 mappers and 10
	// reducers.
	numMappers := 100
	keyValuePairsPerMapper := 1000
	valuesPerKey := 4
	numReducers := 10
	cmd := exec.Command("kelda", "ssh", "spark-driver", "run-example",
		fmt.Sprintf("GroupByTest %d %d %d %d",
			numMappers, keyValuePairsPerMapper, valuesPerKey, numReducers))
	stderrBytes := bytes.NewBuffer(nil)
	cmd.Stderr = stderrBytes

	stdoutBytes, err := cmd.Output()
	fmt.Println(stderrBytes.String())
	fmt.Println(string(stdoutBytes))
	if err != nil {
		t.Fatalf("Failed to run spark job: %s", err.Error())
	}

	// Check for exceptions (this catches a case where a task failed, but then was
	// re-run successfully on a different executor).
	if strings.Contains(stderrBytes.String(), "Exception") ||
		strings.Contains(stderrBytes.String(), "Lost task") {
		t.Fatalf("Exception or lost task found in Spark job output; no " +
			"exceptions or task failures should occur during this test. " +
			"Run tests with -v to see job output.")
	}

	// Make sure that the job ran on multiple machines, instead of just on the
	// master, by checking for log messages that contain "localhost" (which will
	// be present when the job runs in local mode). This test assumes that INFO-level
	// logging is enabled, because the potential log messages containing "localhost"
	// are at INFO level.
	if strings.Contains(stderrBytes.String(), "localhost") {
		t.Fatalf("Spark job expected to run in distributed mode, but it ran " +
			"in local mode")
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
