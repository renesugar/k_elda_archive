package supervisor

import (
	"net"
	"strings"
	"testing"

	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/docker"

	"github.com/stretchr/testify/assert"
)

type testCtx struct {
	fd    fakeDocker
	execs [][]string

	conn db.Conn
}

func initTest() *testCtx {
	conn = db.New()
	md, _dk := docker.NewMock()
	ctx := testCtx{fakeDocker{_dk, md}, nil, conn}
	dk = ctx.fd.Client

	ctx.conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		m := view.InsertMinion()
		m.Self = true
		view.Commit(m)
		e := view.InsertEtcd()
		view.Commit(e)
		return nil
	})

	execRun = func(name string, args ...string) ([]byte, error) {
		ctx.execs = append(ctx.execs, append([]string{name}, args...))
		return nil, nil
	}

	cfgGateway = func(name string, ip net.IPNet) error {
		execRun("cfgGateway", ip.String())
		return nil
	}

	cfgOVN = func(myIP, leaderIP string) error {
		execRun("cfgOvn", myIP, leaderIP)
		return nil
	}

	return &ctx
}

type fakeDocker struct {
	docker.Client
	md *docker.MockClient
}

func (f fakeDocker) running() map[string][]string {
	containers, _ := f.List(nil, false)

	res := map[string][]string{}
	for _, c := range containers {
		res[c.Name] = c.Args
	}
	return res
}

// Test when a stopped version of the container to boot already exists.
func TestJoinContainersWithStopped(t *testing.T) {
	ctx := initTest()
	ro := docker.RunOptions{
		Name:  "name",
		Image: "image",
	}

	// Setup the mock docker such that it looks like there was a container with
	// the desired run options, but exited.
	joinContainers([]docker.RunOptions{ro})
	containersWithName, err := dk.List(map[string][]string{
		"name": {ro.Name},
	}, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, containersWithName)
	exitedContainerID := containersWithName[0].ID
	ctx.fd.md.StopContainer(exitedContainerID)

	// Run the join again. The exited container shouldn't affect booting a new
	// copy of the container.
	joinContainers([]docker.RunOptions{ro})

	containers, err := dk.List(nil, true)
	assert.NoError(t, err)

	var foundNewContainer, foundStoppedContainer bool
	for _, c := range containers {
		if c.Name == ro.Name && c.Running {
			foundNewContainer = true
		}

		if strings.HasPrefix(c.Name, ro.Name+"_stopped") && !c.Running &&
			c.ID == exitedContainerID {
			foundStoppedContainer = true
		}
	}

	assert.True(t, foundNewContainer,
		"a new container should be booted with the container name")
	assert.True(t, foundStoppedContainer,
		"the stopped container should still exist, but be renamed")
}
