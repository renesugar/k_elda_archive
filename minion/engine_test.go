package minion

import (
	"sort"
	"testing"
	"time"

	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/join"
	"github.com/quilt/quilt/stitch"
	"github.com/stretchr/testify/assert"
)

const testImage = "alpine"

func TestContainerTxn(t *testing.T) {
	conn := db.New()
	trigg := conn.Trigger(db.ContainerTable).C

	testContainerTxn(t, conn, stitch.Stitch{})
	assert.False(t, fired(trigg))

	stc := stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
		},
	}
	testContainerTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
			{
				ID:      "6e24c8cbeb63dbffcc82730d01b439e2f5085f59",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
					"6e24c8cbeb63dbffcc82730d01b439e2f5085f59",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "0b8a2ed7d14d78a388375025223b05d072bbaec3",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"cat"},
			},
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"0b8a2ed7d14d78a388375025223b05d072bbaec3",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"0b8a2ed7d14d78a388375025223b05d072bbaec3",
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "7a6244b8d2bfa10ee2fcbe6836a0519e116aee31",
				Image:   stitch.Image{Name: "ubuntu"},
				Command: []string{"cat"},
			},
			{
				ID:      "f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"tail"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"7a6244b8d2bfa10ee2fcbe6836a0519e116aee31",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"7a6244b8d2bfa10ee2fcbe6836a0519e116aee31",
					"f133411ac23f45342a7b8b89bbe5e9efd0e711e5",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:      "0b8a2ed7d14d78a388375025223b05d072bbaec3",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"cat"},
			},
			{
				ID:      "d1c9f501efd7a348e54388358c5fe29690fb147d",
				Image:   stitch.Image{Name: "alpine"},
				Command: []string{"cat"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"0b8a2ed7d14d78a388375025223b05d072bbaec3",
					"d1c9f501efd7a348e54388358c5fe29690fb147d",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				Image: stitch.Image{Name: "alpine"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "a",
				IDs: []string{
					"018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				},
			},
		},
	})
	assert.True(t, fired(trigg))

	stc = stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				Image: stitch.Image{Name: "alpine"},
			},
			{
				ID:    "ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				Image: stitch.Image{Name: "alpine"},
			},
		},
		Labels: []stitch.Label{
			{
				Name: "b",
				IDs: []string{
					"018e4ee517d85640d9bf0adb4579d2ac9bd358af",
				},
			},
			{
				Name: "c",
				IDs: []string{
					"ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				},
			},
			{
				Name: "a",
				IDs: []string{
					"018e4ee517d85640d9bf0adb4579d2ac9bd358af",
					"ac4693f0b7fc17aa0e885aa03dc8f7cd6017f496",
				},
			},
		},
	}
	testContainerTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testContainerTxn(t, conn, stc)
	assert.False(t, fired(trigg))
}

func testContainerTxn(t *testing.T, conn db.Conn, stc stitch.Stitch) {
	var containers []db.Container
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		updatePolicy(view, stc.String())
		containers = view.SelectFromContainer(nil)
		return nil
	})

	for _, e := range queryContainers(stc) {
		found := false
		for i, c := range containers {
			if e.StitchID == c.StitchID {
				containers = append(containers[:i], containers[i+1:]...)
				found = true
				break
			}
		}

		assert.True(t, found)
	}

	assert.Empty(t, containers)
}

func TestConnectionTxn(t *testing.T) {
	conn := db.New()
	trigg := conn.Trigger(db.ConnectionTable).C

	testConnectionTxn(t, conn, stitch.Stitch{})
	assert.False(t, fired(trigg))

	stc := stitch.Stitch{
		Connections: []stitch.Connection{
			{From: "a", To: "a", MinPort: 80, MaxPort: 80},
		},
	}
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	stc.Connections = []stitch.Connection{
		{From: "a", To: "a", MinPort: 90, MaxPort: 90},
	}
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	stc.Connections = []stitch.Connection{
		{From: "b", To: "a", MinPort: 90, MaxPort: 90},
		{From: "b", To: "c", MinPort: 90, MaxPort: 90},
		{From: "b", To: "a", MinPort: 100, MaxPort: 100},
		{From: "c", To: "a", MinPort: 101, MaxPort: 101},
	}
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))

	stc.Connections = nil
	testConnectionTxn(t, conn, stc)
	assert.True(t, fired(trigg))

	testConnectionTxn(t, conn, stc)
	assert.False(t, fired(trigg))
}

func testConnectionTxn(t *testing.T, conn db.Conn, stc stitch.Stitch) {
	var connections []db.Connection
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		updatePolicy(view, stc.String())
		connections = view.SelectFromConnection(nil)
		return nil
	})

	exp := stc.Connections
	for _, e := range exp {
		found := false
		for i, c := range connections {
			if e.From == c.From && e.To == c.To && e.MinPort == c.MinPort &&
				e.MaxPort == c.MaxPort {
				connections = append(
					connections[:i], connections[i+1:]...)
				found = true
				break
			}
		}

		assert.True(t, found)
	}

	assert.Empty(t, connections)
}

func fired(c chan struct{}) bool {
	time.Sleep(5 * time.Millisecond)
	select {
	case <-c:
		return true
	default:
		return false
	}
}

func TestPlacementTxn(t *testing.T) {
	conn := db.New()
	checkPlacement := func(stc stitch.Stitch, exp ...db.Placement) {
		var actual db.PlacementSlice
		conn.Txn(db.AllTables...).Run(func(view db.Database) error {
			updatePolicy(view, stc.String())
			actual = db.PlacementSlice(view.SelectFromPlacement(nil))
			return nil
		})

		key := func(plcmIntf interface{}) interface{} {
			plcm := plcmIntf.(db.Placement)
			plcm.ID = 0 // Ignore the Database ID.

			// If it's a container constraint, the order of TargetContainer
			// and OtherContainer doesn't matter. Therefore, we sort the
			// containers IDs so that the assignment is consistent.
			if plcm.OtherContainer != "" {
				ids := []string{plcm.TargetContainer, plcm.OtherContainer}
				sort.Strings(ids)
				plcm.TargetContainer = ids[0]
				plcm.OtherContainer = ids[1]
			}
			return plcm
		}
		_, missing, extra := join.HashJoin(db.PlacementSlice(exp), actual,
			key, key)
		assert.Empty(t, missing)
		assert.Empty(t, extra)
	}

	fooHostname := "foo"
	barHostname := "bar"
	bazHostname := "baz"
	fooID := "fooID"
	barID := "barID"
	bazID := "bazID"
	stc := stitch.Stitch{
		Containers: []stitch.Container{
			{
				Hostname: fooHostname,
				ID:       fooID,
				Image:    stitch.Image{Name: "foo"},
			},
			{
				Hostname: barHostname,
				ID:       barID,
				Image:    stitch.Image{Name: "bar"},
			},
			{
				Hostname: bazHostname,
				ID:       bazID,
				Image:    stitch.Image{Name: "baz"},
			},
		},
	}

	// Machine placement
	stc.Placements = []stitch.Placement{
		{TargetContainerID: "foo", Exclusive: false, Size: "m4.large"},
	}
	checkPlacement(stc,
		db.Placement{
			TargetContainer: "foo",
			Exclusive:       false,
			Size:            "m4.large",
		},
	)

	// Port placement
	stc.Placements = nil
	stc.Connections = []stitch.Connection{
		{From: stitch.PublicInternetLabel, To: fooHostname, MinPort: 80,
			MaxPort: 80},
		{From: stitch.PublicInternetLabel, To: fooHostname, MinPort: 81,
			MaxPort: 81},
	}
	checkPlacement(stc)

	stc.Connections = []stitch.Connection{
		{From: stitch.PublicInternetLabel, To: fooHostname, MinPort: 80,
			MaxPort: 80},
		{From: stitch.PublicInternetLabel, To: barHostname, MinPort: 80,
			MaxPort: 80},
		{From: stitch.PublicInternetLabel, To: barHostname, MinPort: 81,
			MaxPort: 81},
		{From: stitch.PublicInternetLabel, To: bazHostname, MinPort: 81,
			MaxPort: 81},
	}
	checkPlacement(stc,
		db.Placement{
			TargetContainer: fooID,
			OtherContainer:  barID,
			Exclusive:       true,
		},
		db.Placement{
			TargetContainer: barID,
			OtherContainer:  bazID,
			Exclusive:       true,
		},
	)
}

func checkImage(t *testing.T, conn db.Conn, stc stitch.Stitch, exp ...db.Image) {
	var images []db.Image
	conn.Txn(db.AllTables...).Run(func(view db.Database) error {
		updatePolicy(view, stc.String())
		images = view.SelectFromImage(nil)
		return nil
	})

	key := func(intf interface{}) interface{} {
		im := intf.(db.Image)
		im.ID = 0
		return im
	}
	_, lonelyLeft, lonelyRight := join.HashJoin(
		db.ImageSlice(images), db.ImageSlice(exp), key, key)
	assert.Empty(t, lonelyLeft, "unexpected images")
	assert.Empty(t, lonelyRight, "missing images")
}

func TestImageTxn(t *testing.T) {
	t.Parallel()

	// Regular image that isn't built by Quilt.
	checkImage(t, db.New(), stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "475c40d6070969839ba0f88f7a9bd0cc7936aa30",
				Image: stitch.Image{Name: "image"},
			},
		},
	})

	conn := db.New()
	checkImage(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "96189e4ea36c80171fd842ccc4c3438d06061991",
				Image: stitch.Image{Name: "a", Dockerfile: "1"},
			},
			{
				ID:    "c51d206a1414f1fadf5020e5db35feee91410f79",
				Image: stitch.Image{Name: "a", Dockerfile: "1"},
			},
			{
				ID:    "ede1e03efba48e66be3e51aabe03ec77d9f9def9",
				Image: stitch.Image{Name: "b", Dockerfile: "1"},
			},
			{
				ID:    "133c61c61ef4b49ea26717efe0f0468d455fd317",
				Image: stitch.Image{Name: "c"},
			},
		},
	},
		db.Image{
			Name:       "a",
			Dockerfile: "1",
		},
		db.Image{
			Name:       "b",
			Dockerfile: "1",
		},
	)

	// Ensure existing images are preserved.
	conn.Txn(db.ImageTable).Run(func(view db.Database) error {
		img := view.SelectFromImage(func(img db.Image) bool {
			return img.Name == "a"
		})[0]
		img.DockerID = "id"
		view.Commit(img)
		return nil
	})
	checkImage(t, conn, stitch.Stitch{
		Containers: []stitch.Container{
			{
				ID:    "96189e4ea36c80171fd842ccc4c3438d06061991",
				Image: stitch.Image{Name: "a", Dockerfile: "1"},
			},
			{
				ID:    "18c2c81fb48a2a481af58ba5ad6da0e2b244f060",
				Image: stitch.Image{Name: "b", Dockerfile: "2"},
			},
		},
	},
		db.Image{
			Name:       "a",
			Dockerfile: "1",
			DockerID:   "id",
		},
		db.Image{
			Name:       "b",
			Dockerfile: "2",
		},
	)
}
