package command

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"fmt"
	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/util"
	"os"
)

// Stop contains the options for stopping namespaces.
type Stop struct {
	namespace      string
	onlyContainers bool
	force          bool

	connectionHelper
}

// NewStopCommand creates a new Stop command instance.
func NewStopCommand() *Stop {
	return &Stop{}
}

var stopCommands = `kelda stop [NAMESPACE]`

var stopExplanation = `Stop a deployment.

This will free all resources (e.g. VMs) associated with the deployment.

If no namespace is specified, stop the deployment running in the namespace that is
currently tracked by the daemon.

Confirmation is required, but can be skipped with the -f flag.`

// InstallFlags sets up parsing for command line flags.
func (sCmd *Stop) InstallFlags(flags *flag.FlagSet) {
	sCmd.connectionHelper.InstallFlags(flags)

	flags.StringVar(&sCmd.namespace, "namespace", "", "the namespace to stop")
	flags.BoolVar(&sCmd.onlyContainers, "containers", false,
		"only destroy containers")
	flags.BoolVar(&sCmd.force, "f", false, "stop without confirming")

	flags.Usage = func() {
		util.PrintUsageString(stopCommands, stopExplanation, flags)
	}
}

// Parse parses the command line arguments for the stop command.
func (sCmd *Stop) Parse(args []string) error {
	if len(args) > 0 {
		sCmd.namespace = args[0]
	}

	return nil
}

// Run stops the given namespace.
func (sCmd *Stop) Run() int {
	newCluster := blueprint.Blueprint{
		Namespace: sCmd.namespace,
	}

	currDepl, err := getCurrentDeployment(sCmd.client)
	if err != nil && err != errNoBlueprint {
		log.WithError(err).Error("Failed to get current cluster")
		return 1
	}
	if sCmd.namespace == "" {
		newCluster.Namespace = currDepl.Namespace
	}
	if sCmd.onlyContainers {
		if newCluster.Namespace != currDepl.Namespace {
			log.Error("Stopping only containers for a namespace " +
				"not tracked by the remote daemon is not " +
				"currently supported")
			return 1
		}
		newCluster.Machines = currDepl.Machines
	}

	// If the user is stopping the currently tracked namespace, inform the user of
	// the changes that will be made.
	if currDepl.Namespace == newCluster.Namespace {
		// This equality comparison works because any two blueprints that have
		// the exact same data will have exactly the same string representation.
		if currDepl.String() == newCluster.String() {
			fmt.Println("Nothing to stop.")
			return 0
		}

		containersDelta := len(currDepl.Containers) - len(newCluster.Containers)
		machinesDelta := len(currDepl.Machines) - len(newCluster.Machines)
		fmt.Printf("This will stop %s and %s.\n",
			pluralize(containersDelta, "container"),
			pluralize(machinesDelta, "machine"))
	} else {
		fmt.Println("This will stop an unknown number of machines and " +
			"containers.")
	}

	if !sCmd.force {
		shouldStop, err := confirm(os.Stdin, "Continue stopping deployment?")
		if err != nil {
			log.WithError(err).Error("Unable to get user response.")
			return 1
		}

		if !shouldStop {
			fmt.Println("Stop aborted by user.")
			return 0
		}
	}

	if err := sCmd.client.Deploy(newCluster.String()); err != nil {
		log.WithError(err).Error("Unable to stop namespace.")
		return 1
	}

	log.WithField("namespace", sCmd.namespace).Debug("Stopping namespace")
	return 0
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}
