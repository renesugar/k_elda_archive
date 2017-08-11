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
		ImageID:    "imageid",
		Status:     "testing",
		Hostname:   "hostname",
		Command:    []string{"run", "/bin/sh"},
		Env:        fakeMap,
		Created:    fakeTime,
	}

	exp = "Container-1{run test/test run /bin/sh, ImageID: imageid, " +
		"DockerID: DockerID, Minion: Test, StitchID: 1, IP: 1.2.3.4, " +
		"Hostname: hostname, Env: map[test:tester], " +
		"Status: testing, Created: " + fakeTimeString + "}"

	assert.Equal(t, exp, c.String())
}

func TestContainerHelpers(t *testing.T) {
	conn := New()
	conn.Txn(AllTables...).Run(func(view Database) error {
		c := view.InsertContainer()
		c.IP = "foo"
		view.Commit(c)
		return nil
	})

	dbcs := conn.SelectFromContainer(func(c Container) bool {
		return true
	})
	assert.Len(t, dbcs, 1)
	assert.Equal(t, "foo", dbcs[0].IP)
}
