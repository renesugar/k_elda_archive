package main

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteReadData(t *testing.T) {
	// Location in HDFS to write to.
	hdfsFilepath := "/README.txt"

	// Location of the file to write.  Use the Hadoop readme because it is a
	// text file that conveniently already exists on the container.
	readmeLocalFilepath := "/hadoop/README.txt"

	// Write the Hadoop readme to a location in HDFS.
	writeCmd := exec.Command("kelda", "ssh", "hdfs-namenode", "hdfs",
		"dfs", "-copyFromLocal", readmeLocalFilepath, hdfsFilepath)
	fmt.Println("Writing file to HDFS...")
	outBytes, err := writeCmd.CombinedOutput()
	fmt.Println(string(outBytes))
	if err != nil {
		t.Fatalf("Failed to write to HDFS: %s", err.Error())
	}

	// Make sure the write succeeded by reading the same file, and verifying that
	// the result is the same as what was written.
	resultFilepath := "/README_fromHDFS.txt"
	readCmd := exec.Command("kelda", "ssh", "hdfs-namenode", "hdfs",
		"dfs", "-copyToLocal", hdfsFilepath, resultFilepath)
	fmt.Println("Reading file from HDFS...")
	outBytes, err = readCmd.CombinedOutput()
	fmt.Println(string(outBytes))
	if err != nil {
		t.Fatalf("Failed to read from HDFS: %s", err.Error())
	}

	diffCmd := exec.Command("kelda", "ssh", "hdfs-namenode", "diff", resultFilepath,
		readmeLocalFilepath)
	fmt.Println("Computing the diff between the file written and the file read...")
	outBytes, err = diffCmd.CombinedOutput()
	fmt.Println(string(outBytes))
	if err != nil {
		t.Fatalf("Failed to compute file diff: %s", err.Error())
	}

	// Make sure there was no difference between the files (so the diff should be
	// empty).
	assert.Empty(t, string(outBytes))
}
