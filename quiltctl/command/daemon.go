package command

import (
	"flag"

	"github.com/quilt/quilt/api/server"
	"github.com/quilt/quilt/cluster"
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/engine"
	"github.com/quilt/quilt/util"
	"github.com/quilt/quilt/version"

	log "github.com/Sirupsen/logrus"
)

// Daemon contains the options for running the Quilt daemon.
type Daemon struct {
	*connectionFlags
}

// NewDaemonCommand creates a new Daemon command instance.
func NewDaemonCommand() *Daemon {
	return &Daemon{
		connectionFlags: &connectionFlags{},
	}
}

var daemonCommands = "quilt daemon [-H=<daemon_host>]"
var daemonExplanation = "`daemon` starts the quilt daemon, which listens for quilt " +
	"API requests."

// InstallFlags sets up parsing for command line flags
func (dCmd *Daemon) InstallFlags(flags *flag.FlagSet) {
	dCmd.connectionFlags.InstallFlags(flags)
	flags.Usage = func() {
		util.PrintUsageString(daemonCommands, daemonExplanation, flags)
	}
}

// Parse parses the command line arguments for the daemon command.
func (dCmd *Daemon) Parse(args []string) error {
	return nil
}

// BeforeRun makes any necessary post-parsing transformations.
func (dCmd *Daemon) BeforeRun() error {
	return nil
}

// AfterRun performs any necessary post-run cleanup.
func (dCmd *Daemon) AfterRun() error {
	return nil
}

// Run starts the daemon.
func (dCmd *Daemon) Run() int {
	log.WithField("version", version.Version).Info("Starting Quilt daemon")
	conn := db.New()
	go engine.Run(conn)
	go server.Run(conn, dCmd.host, true)
	cluster.Run(conn)
	return 0
}
