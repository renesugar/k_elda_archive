// Package scheduler is respnosible for deciding on which minion to place each container
// in the cluster.  It does this by updating each container in the Database with the
// PrivateIP of the minion it's assigned to, or the empty string if no assignment could
// be made.  Worker nodes then read these assignments form Etcd, and boot the containers
// that they are instructed to.
package scheduler

import (
	"time"

	"github.com/kelda/kelda/counter"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/docker"
	"github.com/kelda/kelda/minion/network/plugin"
	"github.com/kelda/kelda/util"
	log "github.com/sirupsen/logrus"
)

var c = counter.New("Scheduler")

// Run blocks implementing the scheduler module.
func Run(conn db.Conn, dk docker.Client) {
	bootWait(conn)

	err := dk.ConfigureNetwork(plugin.NetworkName)
	if err != nil {
		log.WithError(err).Fatal("Failed to configure network plugin")
	}

	loopLog := util.NewEventTimer("Scheduler")
	trig := conn.TriggerTick(60, db.MinionTable, db.ContainerTable,
		db.PlacementTable, db.EtcdTable, db.ImageTable).C
	for range trig {
		loopLog.LogStart()
		minion := conn.MinionSelf()

		if minion.Role == db.Worker {
			runWorker(conn, dk, minion.PrivateIP, minion.PublicIP)
		} else if minion.Role == db.Master {
			runMaster(conn)
		}
		loopLog.LogEnd()
	}
}

func bootWait(conn db.Conn) {
	waitFn := isMasterReady
	if conn.MinionSelf().Role == db.Worker {
		waitFn = isWorkerReady
	}

	err := util.BackoffWaitFor(func() bool {
		return waitFn(conn)
	}, 30*time.Second, 1*time.Hour)
	if err != nil {
		panic("timed out waiting for scheduler module to start")
	}
}
