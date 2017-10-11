package scheduler

import (
	"errors"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/docker"
	"github.com/kelda/kelda/minion/network/openflow"
)

func TestRunWorker(t *testing.T) {
	t.Parallel()

	replaceFlows = func(ofcs []openflow.Container) error { return errors.New("err") }

	md, dk := docker.NewMock()
	conn := db.New()
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		container := view.InsertContainer()
		container.Image = "Image"
		container.Minion = "1.2.3.4"
		container.IP = "10.0.0.2"
		view.Commit(container)

		m := view.InsertMinion()
		m.Self = true
		m.PrivateIP = "1.2.3.4"
		view.Commit(m)
		return nil
	})

	// Wrong Minion IP, should do nothing.
	runWorker(conn, dk, "1.2.3.5")
	dkcs, err := dk.List(nil)
	assert.NoError(t, err)
	assert.Len(t, dkcs, 0)

	// Run with a list error, should do nothing.
	md.ListError = true
	runWorker(conn, dk, "1.2.3.4")
	md.ListError = false
	dkcs, err = dk.List(nil)
	assert.NoError(t, err)
	assert.Len(t, dkcs, 0)

	runWorker(conn, dk, "1.2.3.4")
	dkcs, err = dk.List(nil)
	assert.NoError(t, err)
	assert.Len(t, dkcs, 1)
	assert.Equal(t, "Image", dkcs[0].Image)
}

func runSync(dk docker.Client, dbcs []db.Container,
	dkcs []docker.Container) []db.Container {

	changes, tdbcs, tdkcs := syncWorker(dbcs, dkcs)
	doContainers(dk, tdkcs, dockerKill)
	doContainers(dk, tdbcs, dockerRun)
	return changes
}

// This test verifies that the worker synchronizes with the database correctly,
// by booting or destroying containers as appropriate.
func TestSyncWorker(t *testing.T) {
	t.Parallel()

	md, dk := docker.NewMock()

	// The containers that should be running, according to the database. We
	// populate this manually here to simulate different states of the
	// database.
	dbcs := []db.Container{
		{
			ID:      1,
			Image:   "Image1",
			Command: []string{"Cmd1"},
			Env:     map[string]string{"Env": "1"},
		},
	}

	// Test when there are no containers running, but one specified in the
	// database. We should attempt to start the container, and there should be
	// no matching containers. However, the container never starts because
	// we mock an error when starting.
	md.StartError = true
	changed := runSync(dk, dbcs, nil)
	md.StartError = false
	assert.Len(t, changed, 0)

	// The same case as above, except there is no error when starting, so the
	// container should actually get booted.
	runSync(dk, dbcs, nil)

	// The previous test booted the desired container. Therefore, this sync
	// should pair the running container with the desired container.
	dkcs, err := dk.List(nil)
	changed, _, _ = syncWorker(dbcs, dkcs)
	assert.NoError(t, err)

	if changed[0].DockerID != dkcs[0].ID {
		t.Error(spew.Sprintf("Incorrect DockerID: %v", changed))
	}

	// Assert that the pairing specified in `changed` is consistent with the
	// desired container in the database.
	dbcs[0].DockerID = dkcs[0].ID
	assert.Equal(t, dbcs, changed)

	// Ensure that the booted container has the attributes specified in the
	// database.
	assert.Equal(t, dbcs[0].Image, dkcs[0].Image)
	assert.Equal(t, dbcs[0].Command, dkcs[0].Args)
	assert.Equal(t, dbcs[0].Env, dkcs[0].Env)

	// Unassign the DockerID, and run the sync again. Even though the DockerID
	// was unassigned, new containers shouldn't be booted.
	dbcs[0].DockerID = ""
	changed = runSync(dk, dbcs, dkcs)

	// Ensure that runSync did not boot any new containers.
	newDkcs, err := dk.List(nil)
	assert.NoError(t, err)
	assert.Equal(t, dkcs, newDkcs)

	// Assert that the pairing specified in `changed` is consistent with the
	// desired container in the database.
	dbcs[0].DockerID = dkcs[0].ID
	assert.Equal(t, dbcs, changed)

	// Change the desired containers to be empty. Any running containers should
	// be stopped. However, the running container is not actually removed
	// because we mock an error during Remove.
	md.RemoveError = true
	changed = runSync(dk, nil, dkcs)
	md.RemoveError = false
	assert.Len(t, changed, 0)

	// Assert that the running containers has not changed.
	newDkcs, err = dk.List(nil)
	assert.NoError(t, err)
	assert.Equal(t, dkcs, newDkcs)

	// The same case as above, except don't throw an error when removing
	// containers. No containers should be running, and no containers should
	// be paired.
	changed = runSync(dk, nil, dkcs)
	assert.Len(t, changed, 0)

	dkcs, err = dk.List(nil)
	assert.NoError(t, err)
	assert.Len(t, dkcs, 0)
}

func TestInitsFiles(t *testing.T) {
	t.Parallel()

	md, dk := docker.NewMock()
	fileMap := map[string]string{"File": "Contents"}
	dbcs := []db.Container{
		{
			ID:                1,
			Image:             "Image1",
			FilepathToContent: fileMap,
		},
	}

	runSync(dk, dbcs, nil)
	dkcs, err := dk.List(nil)
	assert.NoError(t, err)
	assert.Len(t, dkcs, 1)
	assert.Equal(t, filesHash(fileMap), dkcs[0].Labels[filesKey])
	assert.Equal(t, map[docker.UploadToContainerOptions]struct{}{
		{
			ContainerID: dkcs[0].ID,
			UploadPath:  ".",
			TarPath:     "File",
			Contents:    "Contents",
		}: {},
	}, md.Uploads)
}

func TestSyncJoinScore(t *testing.T) {
	t.Parallel()

	dbc := db.Container{
		IP:                "1.2.3.4",
		Image:             "Image",
		Command:           []string{"cmd"},
		Env:               map[string]string{"a": "b"},
		FilepathToContent: map[string]string{"c": "d"},
		DockerID:          "DockerID",
	}
	dkc := docker.Container{
		IP:     "1.2.3.4",
		Image:  dbc.Image,
		Args:   dbc.Command,
		Env:    dbc.Env,
		Labels: map[string]string{filesKey: filesHash(dbc.FilepathToContent)},
		ID:     dbc.DockerID,
	}

	score := syncJoinScore(dbc, dkc)
	assert.Zero(t, score)

	dbc.Image = "Image1"
	score = syncJoinScore(dbc, dkc)
	assert.Equal(t, -1, score)

	dbc.Image = dkc.Image
	score = syncJoinScore(dbc, dkc)
	assert.Zero(t, score)

	dbc.Command = []string{"wrong"}
	score = syncJoinScore(dbc, dkc)
	assert.Equal(t, -1, score)

	dbc.Command = dkc.Args
	score = syncJoinScore(dbc, dkc)
	assert.Zero(t, score)

	dbc.IP = "wrong"
	score = syncJoinScore(dbc, dkc)
	assert.Equal(t, -1, score)

	dbc.IP = dkc.IP
	score = syncJoinScore(dbc, dkc)
	assert.Zero(t, score)

	dbc.Command = dkc.Args
	dbc.Env = map[string]string{"a": "wrong"}
	score = syncJoinScore(dbc, dkc)
	assert.Equal(t, -1, score)
	dbc.Env = dkc.Env

	dbc.FilepathToContent = map[string]string{"c": "wrong"}
	score = syncJoinScore(dbc, dkc)
	assert.Equal(t, -1, score)

	dbc.FilepathToContent = map[string]string{"c": "d"}
	score = syncJoinScore(dbc, dkc)
	assert.Zero(t, score)

	dkc.ImageID = "id"
	dbc.Command = dkc.Args
	dbc.Env = dkc.Env
	dbc.ImageID = dkc.ImageID
	score = syncJoinScore(dbc, dkc)
	assert.Zero(t, score)

	dbc.ImageID = "wrong"
	score = syncJoinScore(dbc, dkc)
	assert.Equal(t, -1, score)
}

func TestOpenFlowContainers(t *testing.T) {
	conns := []db.Connection{
		{MinPort: 1, MaxPort: 1000},
		{MinPort: 2, MaxPort: 2, From: blueprint.PublicInternetLabel, To: "red"},
		{MinPort: 3, MaxPort: 3, To: blueprint.PublicInternetLabel, From: "red"},
		{MinPort: 4, MaxPort: 4, To: blueprint.PublicInternetLabel, From: "blue"}}

	res := openflowContainers([]db.Container{
		{EndpointID: "f", IP: "1.2.3.4", Hostname: "red"}},
		conns)
	exp := []openflow.Container{{
		Veth:    "f",
		Patch:   "q_f",
		IP:      "1.2.3.4",
		Mac:     "02:00:01:02:03:04",
		ToPub:   map[int]struct{}{3: {}},
		FromPub: map[int]struct{}{2: {}},
	}}
	assert.Equal(t, exp, res)
}
