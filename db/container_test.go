package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContainerString(t *testing.T) {
	c := Container{}
	got := c.String()
	exp := "Container-0{run }"
	assert.Equal(t, got, exp)

	fakeMap := make(map[string]string)
	fakeMap["test"] = "tester"
	fakeTime := time.Now()
	fakeTimeString := fakeTime.String()

	c = Container{
		ID:         1,
		IP:         "1.2.3.4",
		Minion:     "Test",
		EndpointID: "TestEndpoint",
		StitchID:   "1",
		DockerID:   "DockerID",
		Image:      "test/test",
		Status:     "testing",
		Hostname:   "hostname",
		Command:    []string{"run", "/bin/sh"},
		Labels:     []string{"label1"},
		Env:        fakeMap,
		Created:    fakeTime,
	}

	exp = "Container-1{run test/test run /bin/sh, DockerID: DockerID, " +
		"Minion: Test, StitchID: 1, IP: 1.2.3.4, Hostname: hostname, " +
		"Labels: [label1], Env: map[test:tester], Status: testing, " +
		"Created: " + fakeTimeString + "}"

	assert.Equal(t, exp, c.String())
}
