package cloud

import (
	"errors"

	"github.com/quilt/quilt/cloud/foreman"
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/util"

	log "github.com/Sirupsen/logrus"
)

func updateMachineStatuses(conn db.Conn) {
	dbTrig := conn.TriggerTick(30, db.MachineTable).C
	for range util.JoinNotifiers(dbTrig, foreman.ConnectionTrigger) {
		updateMachineStatusesOnce(conn)
	}
}

func updateMachineStatusesOnce(conn db.Conn) {
	conn.Txn(db.MachineTable).Run(func(view db.Database) error {
		for _, dbm := range view.SelectFromMachine(nil) {
			// Don't touch machines that are booting. `clst.boot` will take
			// care of unsetting the status when it's no longer booting.
			if dbm.Status == db.Booting {
				continue
			}

			if newStatus, ok := status(dbm); ok && newStatus != dbm.Status {
				dbm.Status = newStatus
				view.Commit(dbm)
			}
		}
		return nil
	})
}

// status returns a status string for the given machine. If no string could be
// determined, the second return value is false.
func status(m db.Machine) (string, bool) {
	// "Connected" takes priority over other statuses.
	connected := m.PublicIP != "" && isConnected(m.PublicIP)
	if connected {
		return db.Connected, true
	}

	// If we had previously connected, and we are not currently connected, show
	// that we are attempting to reconnect.
	if m.Status == db.Connected || m.Status == db.Reconnecting {
		return db.Reconnecting, true
	}

	// If we've never successfully connected, but have booted enough to have a
	// public IP, show that we are attempting to connect.
	if m.PublicIP != "" {
		return db.Connecting, true
	}

	return "", false
}

func setStatuses(conn db.Conn, machines []db.Machine, status string) {
	for _, m := range machines {
		if err := setStatus(conn, m.StitchID, status); err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"machine": m.StitchID,
				"status":  status,
			}).Warn("Failed to update status of machine")
		}
	}
}

func setStatus(conn db.Conn, id string, status string) error {
	return conn.Txn(db.MachineTable).Run(func(view db.Database) error {
		matchingMachines := view.SelectFromMachine(func(m db.Machine) bool {
			return m.StitchID == id
		})
		switch len(matchingMachines) {
		case 1:
			matchingMachines[0].Status = status
			view.Commit(matchingMachines[0])
			return nil
		case 0:
			return errors.New("no matching machine")
		default:
			return errors.New("multiple matching machines")
		}
	})
}

var isConnected = foreman.IsConnected
