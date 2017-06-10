// +build !windows

package minion

import (
	"fmt"
	"time"

	"github.com/quilt/quilt/api"
	apiServer "github.com/quilt/quilt/api/server"
	"github.com/quilt/quilt/connection"
	"github.com/quilt/quilt/counter"
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/minion/docker"
	"github.com/quilt/quilt/minion/etcd"
	"github.com/quilt/quilt/minion/network"
	"github.com/quilt/quilt/minion/network/plugin"
	"github.com/quilt/quilt/minion/pprofile"
	"github.com/quilt/quilt/minion/registry"
	"github.com/quilt/quilt/minion/scheduler"
	"github.com/quilt/quilt/minion/supervisor"
	"github.com/quilt/quilt/quiltctl/command/credentials"
	"github.com/quilt/quilt/util"

	log "github.com/Sirupsen/logrus"
)

var c = counter.New("Minion")

// Run blocks executing the minion.
func Run(role db.Role, inboundPubIntf, outboundPubIntf, tlsDir string) {
	// XXX Uncomment the following line to run the profiler
	//runProfiler(5 * time.Minute)

	conn := db.New()
	dk := docker.New("unix:///var/run/docker.sock")

	// XXX: As we are developing minion modules to use this passed down role
	// instead of querying their db independently, we need to do this.
	// Possibly in the future just pass down role into all of the modules,
	// but may be simpler to just have it use this entry.
	conn.Txn(db.MinionTable).Run(func(view db.Database) error {
		minion := view.InsertMinion()
		minion.Role = role
		minion.Self = true
		view.Commit(minion)
		return nil
	})

	if role == db.Worker {
		// Start writing the machine's subnets as soon as possible so that the
		// master can make informed IP allocations.
		go network.WriteSubnets(conn)
	}

	// Not in a goroutine, want the plugin to start before the scheduler
	plugin.Run()

	supervisor.Run(conn, dk, role)

	go scheduler.Run(conn, dk)
	go network.Run(conn, inboundPubIntf, outboundPubIntf)
	go registry.Run(conn, dk)
	go etcd.Run(conn)
	go syncAuthorizedKeys(conn)

	// Block until the credentials are in place on the local filesystem. We
	// can't simply fail if the first read fails because the daemon might still
	// be generating and copying keys onto the local filesystem. The key
	// installation is handled by SyncCredentials in cluster/credentials.go.
	var creds connection.Credentials
	err := util.BackoffWaitFor(func() bool {
		var err error
		creds, err = credentials.Read(tlsDir)
		if err != nil {
			log.WithError(err).Debug("TLS keys not ready yet")
			return false
		}
		return true
	}, 30*time.Second, 1*time.Hour)
	if err != nil {
		log.Error("Failed to read minion credentials")
		return
	}

	go minionServerRun(conn, creds)
	go apiServer.Run(conn, fmt.Sprintf("tcp://0.0.0.0:%d", api.DefaultRemotePort),
		false, creds)

	loopLog := util.NewEventTimer("Minion-Update")

	for range conn.Trigger(db.MinionTable, db.EtcdTable).C {
		loopLog.LogStart()
		txn := conn.Txn(db.ConnectionTable, db.ContainerTable, db.MinionTable,
			db.EtcdTable, db.PlacementTable, db.ImageTable)
		txn.Run(func(view db.Database) error {
			minion := view.MinionSelf()
			if view.EtcdLeader() {
				updatePolicy(view, minion.Blueprint)
			}
			return nil
		})
		loopLog.LogEnd()
	}
}

func runProfiler(duration time.Duration) {
	go func() {
		p := pprofile.New("minion")
		for {
			if err := p.TimedRun(duration); err != nil {
				log.Error(err)
			}
		}
	}()
}
