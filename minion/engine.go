package minion

import (
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/join"
	"github.com/quilt/quilt/stitch"

	log "github.com/Sirupsen/logrus"
)

func updatePolicy(view db.Database, blueprint string) {
	compiled, err := stitch.FromJSON(blueprint)
	if err != nil {
		log.WithError(err).Warn("Invalid blueprint.")
		return
	}

	c.Inc("Update Policy")
	updateImages(view, compiled)
	updateContainers(view, compiled)
	updateLabels(view, compiled)
	updateConnections(view, compiled)
	updatePlacements(view, compiled)
}

// `portPlacements` creates exclusive placement rules such that no two containers
// listening on the same public port get placed on the same machine.
func portPlacements(connections []db.Connection, containers []db.Container) (
	placements []db.Placement) {

	hostnameToContainer := map[string]db.Container{}
	for _, c := range containers {
		hostnameToContainer[c.Hostname] = c
	}

	ports := make(map[int][]string)
	for _, conn := range connections {
		if conn.From != stitch.PublicInternetLabel {
			continue
		}

		toContainer, ok := hostnameToContainer[conn.To]
		if !ok {
			log.WithField("connection", conn).
				WithField("hostname", conn.To).
				Warn("Public connection in terms of unknown hostname." +
					"Ignoring.")
			continue
		}

		// XXX: Public connections do not currently support ranges, so we can
		// safely consider just the MinPort.
		ports[conn.MinPort] = append(ports[conn.MinPort], toContainer.StitchID)
	}

	// Create placement rules for all combinations of containers that listen on
	// the same port. We do not need to create a rule for every permutation
	// because order does not matter for the `TargetContainer` and
	// `OtherContainer` fields -- the placement is equivalent if the two fields
	// are swapped.  We do so by creating a placement rule between each
	// container, and the containers after it. There is no need to create rules
	// for the preceding containers because the previous rules will have
	// covered it.
	for _, cids := range ports {
		for i, tgt := range cids {
			for _, other := range cids[i+1:] {
				placements = append(placements,
					db.Placement{
						Exclusive:       true,
						TargetContainer: tgt,
						OtherContainer:  other,
					},
				)
			}
		}
	}

	return placements
}

func updatePlacements(view db.Database, blueprint stitch.Stitch) {
	connections := view.SelectFromConnection(nil)
	containers := view.SelectFromContainer(nil)
	placements := db.PlacementSlice(portPlacements(connections, containers))
	for _, sp := range blueprint.Placements {
		placements = append(placements, db.Placement{
			TargetContainer: sp.TargetContainerID,
			Exclusive:       sp.Exclusive,
			Provider:        sp.Provider,
			Size:            sp.Size,
			Region:          sp.Region,
			FloatingIP:      sp.FloatingIP,
		})
	}

	key := func(val interface{}) interface{} {
		p := val.(db.Placement)
		p.ID = 0
		return p
	}

	dbPlacements := db.PlacementSlice(view.SelectFromPlacement(nil))
	_, addSet, removeSet := join.HashJoin(placements, dbPlacements, key, key)

	for _, toAddIntf := range addSet {
		toAdd := toAddIntf.(db.Placement)

		id := view.InsertPlacement().ID
		toAdd.ID = id
		view.Commit(toAdd)
	}

	for _, toRemove := range removeSet {
		view.Remove(toRemove.(db.Placement))
	}
}

func updateLabels(view db.Database, blueprint stitch.Stitch) {
	var stitchLabels db.LabelSlice
	for _, label := range blueprint.Labels {
		stitchLabels = append(stitchLabels, db.Label{
			Label:     label.Name,
			Hostnames: label.Hostnames,
		})
	}

	key := func(intf interface{}) interface{} {
		return intf.(db.Label).Label
	}

	dbLabels := db.LabelSlice(view.SelectFromLabel(nil))
	pairs, toAdd, toRemove := join.HashJoin(stitchLabels, dbLabels,
		key, key)

	for _, intf := range toRemove {
		view.Remove(intf.(db.Label))
	}

	for _, intf := range toAdd {
		pairs = append(pairs, join.Pair{L: intf, R: view.InsertLabel()})
	}

	for _, pair := range pairs {
		dbLabel := pair.R.(db.Label)
		stitchLabel := pair.L.(db.Label)

		// Modify the original database label so that we preserve whatever IP
		// the label might have already been allocated.
		dbLabel.Label = stitchLabel.Label
		dbLabel.Hostnames = stitchLabel.Hostnames
		view.Commit(dbLabel)
	}
}

func updateConnections(view db.Database, blueprint stitch.Stitch) {
	scs := stitch.ConnectionSlice(blueprint.Connections)

	// Setup connections to load balanced containers. Load balancing works by
	// rewriting the load balancer IPs to the IP address of one of the load
	// balanced containers. This means allowing connections only to the load
	// balancer IP address is insufficient -- the container must also be able
	// to communicate directly with the containers behind the load balancer.
	labels := map[string]stitch.Label{}
	for _, label := range blueprint.Labels {
		labels[label.Name] = label
	}

	for _, c := range scs {
		label, ok := labels[c.To]
		if !ok {
			continue
		}

		for _, hostname := range label.Hostnames {
			scs = append(scs, stitch.Connection{
				From:    c.From,
				To:      hostname,
				MinPort: c.MinPort,
				MaxPort: c.MaxPort,
			})
		}
	}

	dbcKey := func(val interface{}) interface{} {
		c := val.(db.Connection)
		return stitch.Connection{
			From:    c.From,
			To:      c.To,
			MinPort: c.MinPort,
			MaxPort: c.MaxPort,
		}
	}

	vcs := view.SelectFromConnection(nil)
	pairs, stitches, dbcs := join.HashJoin(scs, db.ConnectionSlice(vcs), nil, dbcKey)

	for _, dbc := range dbcs {
		view.Remove(dbc.(db.Connection))
	}

	for _, stitchc := range stitches {
		pairs = append(pairs, join.Pair{L: stitchc, R: view.InsertConnection()})
	}

	for _, pair := range pairs {
		stitchc := pair.L.(stitch.Connection)
		dbc := pair.R.(db.Connection)

		dbc.From = stitchc.From
		dbc.To = stitchc.To
		dbc.MinPort = stitchc.MinPort
		dbc.MaxPort = stitchc.MaxPort
		view.Commit(dbc)
	}
}

func queryContainers(blueprint stitch.Stitch) []db.Container {
	containers := map[string]*db.Container{}
	for _, c := range blueprint.Containers {
		containers[c.Hostname] = &db.Container{
			StitchID:          c.ID,
			Command:           c.Command,
			Env:               c.Env,
			FilepathToContent: c.FilepathToContent,
			Image:             c.Image.Name,
			Dockerfile:        c.Image.Dockerfile,
			Hostname:          c.Hostname,
		}
	}

	var ret []db.Container
	for _, c := range containers {
		ret = append(ret, *c)
	}

	return ret
}

func updateContainers(view db.Database, blueprint stitch.Stitch) {
	key := func(val interface{}) interface{} {
		return val.(db.Container).StitchID
	}

	pairs, news, dbcs := join.HashJoin(db.ContainerSlice(queryContainers(blueprint)),
		db.ContainerSlice(view.SelectFromContainer(nil)), key, key)

	for _, dbc := range dbcs {
		view.Remove(dbc.(db.Container))
	}

	for _, new := range news {
		pairs = append(pairs, join.Pair{L: new, R: view.InsertContainer()})
	}

	for _, pair := range pairs {
		newc := pair.L.(db.Container)
		dbc := pair.R.(db.Container)

		dbc.Command = newc.Command
		dbc.Image = newc.Image
		dbc.Dockerfile = newc.Dockerfile
		dbc.Env = newc.Env
		dbc.FilepathToContent = newc.FilepathToContent
		dbc.StitchID = newc.StitchID
		dbc.Hostname = newc.Hostname
		view.Commit(dbc)
	}
}

func updateImages(view db.Database, blueprint stitch.Stitch) {
	dbImageKey := func(intf interface{}) interface{} {
		return stitch.Image{
			Name:       intf.(db.Image).Name,
			Dockerfile: intf.(db.Image).Dockerfile,
		}
	}

	blueprintImages := stitchImageSlice(queryImages(blueprint))
	dbImages := db.ImageSlice(view.SelectFromImage(nil))
	_, toAdd, toRemove := join.HashJoin(blueprintImages, dbImages, nil, dbImageKey)

	for _, intf := range toAdd {
		im := view.InsertImage()
		im.Name = intf.(stitch.Image).Name
		im.Dockerfile = intf.(stitch.Image).Dockerfile
		view.Commit(im)
	}

	for _, row := range toRemove {
		view.Remove(row.(db.Image))
	}
}

func queryImages(blueprint stitch.Stitch) (images []stitch.Image) {
	addedImages := map[stitch.Image]struct{}{}
	for _, c := range blueprint.Containers {
		_, addedImage := addedImages[c.Image]
		if c.Image.Dockerfile == "" || addedImage {
			continue
		}

		images = append(images, c.Image)
		addedImages[c.Image] = struct{}{}
	}
	return images
}

type stitchImageSlice []stitch.Image

func (slc stitchImageSlice) Get(ii int) interface{} {
	return slc[ii]
}

func (slc stitchImageSlice) Len() int {
	return len(slc)
}
