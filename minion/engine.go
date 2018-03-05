package minion

import (
	"fmt"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/join"
	"github.com/kelda/kelda/util"
	"github.com/kelda/kelda/util/str"

	log "github.com/sirupsen/logrus"
)

// updatePolicyTables are the tables required for the `updatePolicy` function
// to run.
var updatePolicyTables = []db.TableType{db.BlueprintTable, db.ConnectionTable,
	db.ContainerTable, db.EtcdTable, db.PlacementTable, db.ImageTable,
	db.LoadBalancerTable}

func syncPolicy(conn db.Conn) {
	loopLog := util.NewEventTimer("Minion-Update")
	for range conn.Trigger(db.EtcdTable).C {
		loopLog.LogStart()
		conn.Txn(updatePolicyTables...).Run(func(view db.Database) error {
			updatePolicy(view)
			return nil
		})
		loopLog.LogEnd()
	}
}

func updatePolicy(view db.Database) {
	if !view.EtcdLeader() {
		return
	}

	bpRow, err := view.GetBlueprint()
	if err != nil {
		log.WithError(err).Error("Failed to get blueprint.")
		return
	}
	bp := bpRow.Blueprint

	c.Inc("Update Policy")
	updateImages(view, bp)
	updateContainers(view, bp)
	updateLoadBalancers(view, bp)
	updateConnections(view, bp)
	updatePlacements(view, bp)
}

// `portPlacements` creates exclusive placement rules such that no two
// containers listening on the same public port get placed on the same machine.
// It produces the same placement rules regardless of the order of connections
// by creating rules for both containers affected by the placement.
// This way, if `portPlacements` is called multiple times for the same
// blueprint, the placement rules will not change unnecessarily.
func portPlacements(connections []db.Connection) (placements []db.Placement) {
	ports := make(map[int][]string)
	for _, conn := range connections {
		if !str.SliceContains(conn.From, blueprint.PublicInternetLabel) {
			continue
		}

		// XXX: Public connections do not currently support ranges, so we
		// can safely consider just the MinPort.
		ports[conn.MinPort] = append(ports[conn.MinPort],
			str.SliceFilterOut(conn.To, blueprint.PublicInternetLabel)...)
	}

	// Create placement rules for all combinations of containers that listen on
	// the same port.
	for _, cids := range ports {
		for i, tgt := range cids {
			for j, other := range cids {
				if i == j {
					continue
				}

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

func updatePlacements(view db.Database, bp blueprint.Blueprint) {
	connections := view.SelectFromConnection(nil)
	placements := db.PlacementSlice(portPlacements(connections))
	for _, sp := range bp.Placements {
		placements = append(placements, db.Placement{
			TargetContainer: sp.TargetContainer,
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

func updateLoadBalancers(view db.Database, bp blueprint.Blueprint) {
	var bpLoadBalancers db.LoadBalancerSlice
	for _, lb := range bp.LoadBalancers {
		bpLoadBalancers = append(bpLoadBalancers, db.LoadBalancer{
			Name:      lb.Name,
			Hostnames: lb.Hostnames,
		})
	}

	key := func(intf interface{}) interface{} {
		return intf.(db.LoadBalancer).Name
	}

	dbLoadBalancers := db.LoadBalancerSlice(view.SelectFromLoadBalancer(nil))
	pairs, toAdd, toRemove := join.HashJoin(bpLoadBalancers, dbLoadBalancers,
		key, key)

	for _, intf := range toRemove {
		view.Remove(intf.(db.LoadBalancer))
	}

	for _, intf := range toAdd {
		pairs = append(pairs, join.Pair{L: intf, R: view.InsertLoadBalancer()})
	}

	for _, pair := range pairs {
		dbLoadBalancer := pair.R.(db.LoadBalancer)
		bpLoadBalancer := pair.L.(db.LoadBalancer)

		// Modify the original database load balancer so that we preserve
		// whatever IP the load balancer might have already been allocated.
		dbLoadBalancer.Name = bpLoadBalancer.Name
		dbLoadBalancer.Hostnames = bpLoadBalancer.Hostnames
		view.Commit(dbLoadBalancer)
	}
}

func updateConnections(view db.Database, bp blueprint.Blueprint) {
	scs := blueprint.ConnectionSlice(bp.Connections)

	// Setup connections to load balanced containers. Load balancing works by
	// rewriting the load balancer IPs to the IP address of one of the load
	// balanced containers. This means allowing connections only to the load
	// balancer IP address is insufficient -- the container must also be able
	// to communicate directly with the containers behind the load balancer.
	loadBalancers := map[string]blueprint.LoadBalancer{}
	for _, lb := range bp.LoadBalancers {
		loadBalancers[lb.Name] = lb
	}

	for _, c := range scs {
		for _, to := range c.To {
			if lb, ok := loadBalancers[to]; ok {
				scs = append(scs, blueprint.Connection{
					From:    c.From,
					To:      lb.Hostnames,
					MinPort: c.MinPort,
					MaxPort: c.MaxPort,
				})
			}
		}
	}

	dbcKey := func(val interface{}) interface{} {
		c := val.(db.Connection)
		return fmt.Sprintf("%s %s %d %d", c.From, c.To, c.MinPort, c.MaxPort)
	}

	bpKey := func(val interface{}) interface{} {
		c := val.(blueprint.Connection)
		return fmt.Sprintf("%s %s %d %d", c.From, c.To, c.MinPort, c.MaxPort)
	}

	vcs := view.SelectFromConnection(nil)
	_, bpcs, dbcs := join.HashJoin(scs, db.ConnectionSlice(vcs), bpKey, dbcKey)

	for _, dbc := range dbcs {
		view.Remove(dbc.(db.Connection))
	}

	for _, newbpconn := range bpcs {
		bpc := newbpconn.(blueprint.Connection)
		dbc := view.InsertConnection()

		dbc.From = bpc.From
		dbc.To = bpc.To
		dbc.MinPort = bpc.MinPort
		dbc.MaxPort = bpc.MaxPort
		view.Commit(dbc)
	}
}

func queryContainers(bp blueprint.Blueprint) []db.Container {
	containers := map[string]*db.Container{}
	for _, c := range bp.Containers {
		containers[c.Hostname] = &db.Container{
			BlueprintID:       c.ID,
			Command:           c.Command,
			Env:               c.Env,
			FilepathToContent: c.FilepathToContent,
			Image:             c.Image.Name,
			Dockerfile:        c.Image.Dockerfile,
			Hostname:          c.Hostname,
			Privileged:        c.Privileged,
			VolumeMounts:      c.VolumeMounts,
		}
	}

	var ret []db.Container
	for _, c := range containers {
		ret = append(ret, *c)
	}

	return ret
}

func updateContainers(view db.Database, bp blueprint.Blueprint) {
	key := func(val interface{}) interface{} {
		return val.(db.Container).BlueprintID
	}

	pairs, news, dbcs := join.HashJoin(db.ContainerSlice(queryContainers(bp)),
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
		dbc.BlueprintID = newc.BlueprintID
		dbc.Hostname = newc.Hostname
		dbc.Privileged = newc.Privileged
		dbc.VolumeMounts = newc.VolumeMounts
		view.Commit(dbc)
	}
}

func updateImages(view db.Database, bp blueprint.Blueprint) {
	dbImageKey := func(intf interface{}) interface{} {
		return blueprint.Image{
			Name:       intf.(db.Image).Name,
			Dockerfile: intf.(db.Image).Dockerfile,
		}
	}

	blueprintImages := blueprintImageSlice(queryImages(bp))
	dbImages := db.ImageSlice(view.SelectFromImage(nil))
	_, toAdd, toRemove := join.HashJoin(blueprintImages, dbImages, nil, dbImageKey)

	for _, intf := range toAdd {
		im := view.InsertImage()
		im.Name = intf.(blueprint.Image).Name
		im.Dockerfile = intf.(blueprint.Image).Dockerfile
		view.Commit(im)
	}

	for _, row := range toRemove {
		view.Remove(row.(db.Image))
	}
}

func queryImages(bp blueprint.Blueprint) (images []blueprint.Image) {
	addedImages := map[blueprint.Image]struct{}{}
	for _, c := range bp.Containers {
		_, addedImage := addedImages[c.Image]
		if c.Image.Dockerfile == "" || addedImage {
			continue
		}

		images = append(images, c.Image)
		addedImages[c.Image] = struct{}{}
	}
	return images
}

type blueprintImageSlice []blueprint.Image

func (slc blueprintImageSlice) Get(ii int) interface{} {
	return slc[ii]
}

func (slc blueprintImageSlice) Len() int {
	return len(slc)
}
